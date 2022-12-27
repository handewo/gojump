package handler

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gliderlabs/ssh"
	"github.com/handewo/gojump/pkg/common"
	"github.com/handewo/gojump/pkg/config"
	"github.com/handewo/gojump/pkg/core"
	"github.com/handewo/gojump/pkg/log"
	"github.com/handewo/gojump/pkg/model"
	"github.com/xlab/treeprint"
)

func NewInteractiveHandler(sess ssh.Session, user *model.User, core *core.Core,
	termConfig model.TerminalConfig) *InteractiveHandler {
	wrapperSess := NewWrapperSession(sess)
	term := common.NewTerminal(wrapperSess, "Opt> ")
	handler := &InteractiveHandler{
		sess:         wrapperSess,
		user:         user,
		term:         term,
		core:         core,
		terminalConf: &termConfig,
	}
	handler.Initial()
	return handler
}

func checkMaxIdleTime(maxIdleMinutes int, user *model.User, sess ssh.Session, checkChan <-chan bool) {
	maxIdleTime := time.Duration(maxIdleMinutes) * time.Minute
	tick := time.NewTicker(maxIdleTime)
	defer tick.Stop()
	checkStatus := true
	for {
		select {
		case <-tick.C:
			if checkStatus {
				msg := fmt.Sprintf("Connect idle more than %d minutes, disconnect", maxIdleMinutes)
				_, _ = io.WriteString(sess, "\r\n"+msg+"\r\n")
				_ = sess.Close()
				log.Debug.Printf("User %s input idle more than %d minutes", user.Username, maxIdleMinutes)
			}
		case <-sess.Context().Done():
			log.Debug.Printf("Stop checking user %s input idle time", user.Username)
			return
		case checkStatus = <-checkChan:
			if !checkStatus {
				//log.Debug.Printf("Stop checking user %s idle time if more than %d minutes", user.Username, maxIdleMinutes)
				continue
			}
			tick.Reset(maxIdleTime)
			//log.Debug.Printf("Start checking user %s idle time if more than %d minutes", user.Username, maxIdleMinutes)
		}
	}
}

type InteractiveHandler struct {
	sess *WrapperSession
	user *model.User
	term *common.Terminal

	selectHandler *UserSelectHandler

	nodes model.NodeList

	wg sync.WaitGroup

	core *core.Core

	terminalConf *model.TerminalConfig
}

func (h *InteractiveHandler) Initial() {
	conf := config.GetConf()
	if conf.ClientAliveInterval > 0 {
		go h.keepSessionAlive(time.Duration(conf.ClientAliveInterval) * time.Second)
	}

	h.selectHandler = &UserSelectHandler{
		user:     h.user,
		h:        h,
		pageInfo: &pageInfo{},
	}
	allAssets, nodes, err := h.core.GetAllUserPermsAssets(h.user.NodeIDs)
	if err != nil {
		log.Error.Printf("Get all user perms assets failed: %s", err)
	}
	h.selectHandler.SetAllLocalData(allAssets)
	h.nodes = nodes

}

func (h *InteractiveHandler) displayHelp() {
	h.term.SetPrompt("Opt> ")
	h.displayBanner(h.sess, h.terminalConf)
}

func (h *InteractiveHandler) WatchWinSizeChange(winChan <-chan ssh.Window) {
	defer log.Debug.Printf("Request %s: Windows change watch close", h.sess.Uuid[:8])
	for {
		select {
		case <-h.sess.Sess.Context().Done():
			return
		case win, ok := <-winChan:
			if !ok {
				return
			}
			h.sess.SetWin(win)
			log.Debug.Printf("Term window size change: %d*%d", win.Height, win.Width)
			_ = h.term.SetSize(win.Width, win.Height)
		}
	}
}

func (h *InteractiveHandler) keepSessionAlive(keepAliveTime time.Duration) {
	t := time.NewTicker(keepAliveTime)
	defer t.Stop()
	for {
		select {
		case <-h.sess.Sess.Context().Done():
			return
		case <-t.C:
			_, err := h.sess.Sess.SendRequest("keepalive@openssh.com", true, nil)
			if err != nil {
				log.Error.Printf("Request %s: Send user %s keepalive packet failed: %s",
					h.sess.Uuid[:8], h.user.Username, err)
				continue
			}
			log.Debug.Printf("Request %s: Send user %s keepalive packet success", h.sess.Uuid[:8], h.user.Username)
		}
	}
}

func (h *InteractiveHandler) chooseSystemUser(systemUsers []model.SystemUser) (systemUser model.SystemUser, ok bool) {
	length := len(systemUsers)
	switch length {
	case 0:
		warningInfo := ("No system user found.")
		_, _ = io.WriteString(h.term, warningInfo+"\n\r")
		return model.SystemUser{}, false
	case 1:
		return systemUsers[0], true
	default:
	}

	idLabel := ("ID")
	nameLabel := ("Name")
	usernameLabel := ("Username")

	labels := []string{idLabel, nameLabel, usernameLabel}
	fields := []string{"ID", "Name", "Username"}

	data := make([]map[string]string, len(systemUsers))
	for i, j := range systemUsers {
		row := make(map[string]string)
		row["ID"] = strconv.Itoa(i + 1)
		row["Name"] = j.Username
		row["Username"] = j.Username
		data[i] = row
	}
	w, _ := h.term.GetSize()
	table := common.WrapperTable{
		Fields: fields,
		Labels: labels,
		FieldsSize: map[string][3]int{
			"ID":       {0, 0, 5},
			"Name":     {0, 8, 0},
			"Username": {0, 10, 0},
		},
		Data:        data,
		TotalSize:   w,
		TruncPolicy: common.TruncMiddle,
	}
	table.Initial()

	h.term.SetPrompt("ID> ")
	selectTip := ("Tips: Enter system user ID and directly login")
	backTip := ("Back: B/b")
	for {
		common.IgnoreErrWriteString(h.term, table.Display())
		common.IgnoreErrWriteString(h.term, common.WrapperString(selectTip, common.Green))
		common.IgnoreErrWriteString(h.term, common.CharNewLine)
		common.IgnoreErrWriteString(h.term, common.WrapperString(backTip, common.Green))
		common.IgnoreErrWriteString(h.term, common.CharNewLine)
		line, err := h.term.ReadLine()
		if err != nil {
			return
		}
		line = strings.TrimSpace(line)
		switch strings.ToLower(line) {
		case "q", "b", "quit", "exit", "back":
			return
		}
		if num, err := strconv.Atoi(line); err == nil {
			if num > 0 && num <= len(systemUsers) {
				return systemUsers[num-1], true
			}
		}
	}
}

func (h *InteractiveHandler) refreshAssetsAndNodesData() {
	h.wg.Add(2)
	go func() {
		defer h.wg.Done()
		allAssets, nodes, err := h.core.GetAllUserPermsAssets(h.user.NodeIDs)
		if err != nil {
			log.Error.Printf("Refresh user all perms assets error: %s", err)
			return
		}
		h.selectHandler.SetAllLocalData(allAssets)
		h.nodes = nodes
	}()
	go func() {
		defer h.wg.Done()
		tConfig, err := h.core.GetTerminalConfig()
		if err != nil {
			log.Error.Printf("Refresh user terminal config error: %s", err)
			return
		}
		h.terminalConf = &tConfig
	}()
	h.wg.Wait()
	_, err := io.WriteString(h.term, ("Refresh done")+"\n\r")
	if err != nil {
		log.Error.Print("refresh Assets Nodes err:", err)
	}
}

func getPageSize(term *common.Terminal, termConf *model.TerminalConfig) int {
	var (
		pageSize  int
		minHeight = 8 // 分页显示的最小高度
	)
	_, height := term.GetSize()

	AssetListPageSize := termConf.AssetListPageSize
	switch AssetListPageSize {
	case "auto":
		pageSize = height - minHeight
	case "all":
		return PAGESIZEALL
	default:
		if value, err := strconv.Atoi(AssetListPageSize); err == nil {
			pageSize = value
		} else {
			pageSize = height - minHeight
		}
	}
	if pageSize <= 0 {
		pageSize = 1
	}
	return pageSize
}

func ConstructNodeTree(assetNodes []model.Node) treeprint.Tree {
	model.SortNodesByKey(assetNodes)
	keyIndexMap := make(map[string]int)
	for index := range assetNodes {
		keyIndexMap[assetNodes[index].Key] = index
	}
	rootTree := treeprint.New()
	constructDisplayTree(rootTree, convertToDisplayTrees(assetNodes), keyIndexMap)
	return rootTree
}

func constructDisplayTree(tree treeprint.Tree, rootNodes []*displayTree, keyMap map[string]int) {
	for i := 0; i < len(rootNodes); i++ {
		subTree := tree.AddBranch(fmt.Sprintf("%d.%s(%d)",
			keyMap[rootNodes[i].Key]+1, rootNodes[i].node.Name,
			len(rootNodes[i].node.AssetIDs)))
		if len(rootNodes[i].subTrees) > 0 {
			sort.Sort(nodeTrees(rootNodes[i].subTrees))
			constructDisplayTree(subTree, rootNodes[i].subTrees, keyMap)
		}
	}
}

func convertToDisplayTrees(assetNodes []model.Node) []*displayTree {
	var rootNodeTrees []*displayTree
	nodeTreeMap := make(map[string]*displayTree)
	for i := 0; i < len(assetNodes); i++ {
		currentTree := displayTree{
			Key:  assetNodes[i].Key,
			node: assetNodes[i],
		}
		r := strings.LastIndex(assetNodes[i].Key, ":")
		if r < 0 {
			rootNodeTrees = append(rootNodeTrees, &currentTree)
			nodeTreeMap[assetNodes[i].Key] = &currentTree
			continue
		}
		nodeTreeMap[assetNodes[i].Key] = &currentTree
		parentKey := assetNodes[i].Key[:r]

		parentTree, ok := nodeTreeMap[parentKey]
		if !ok {
			rootNodeTrees = append(rootNodeTrees, &currentTree)
			continue
		}
		parentTree.AddSubNode(&currentTree)
	}
	return rootNodeTrees
}

type displayTree struct {
	Key      string
	node     model.Node
	subTrees []*displayTree
}

func (t *displayTree) AddSubNode(sub *displayTree) {
	t.subTrees = append(t.subTrees, sub)
}

type nodeTrees []*displayTree

func (l nodeTrees) Len() int {
	return len(l)
}

func (l nodeTrees) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (l nodeTrees) Less(i, j int) bool {
	return l[i].node.Name < l[j].node.Name
}
