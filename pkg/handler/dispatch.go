package handler

import (
	"io"
	"strconv"
	"strings"

	"github.com/handewo/gojump/pkg/log"
)

func (h *InteractiveHandler) Dispatch() {
	defer log.Info.Printf("Request %s: User %s stop interactive", h.sess.ID()[:8], h.user.Username)
	var initialed bool
	checkChan := make(chan bool)
	go h.checkMaxIdleTime(checkChan)
	if h.selectHandler.HasJustOneAsset() {
		checkChan <- false
		h.selectHandler.SearchOrProxy("")
		return
	}
	h.displayHelp()
	for {
		checkChan <- true
		line, err := h.term.ReadLine()
		if err != nil {
			log.Debug.Printf("User %s close connect %s", h.user.Username, err)
			break
		}
		checkChan <- false
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			// default action
			if initialed {
				h.selectHandler.MoveNextPage()
			} else {
				h.selectHandler.SetSelectType(TypeAsset)
				h.selectHandler.Search("")
			}
			initialed = true
			continue
		}
		initialed = true
		switch len(line) {
		case 1:
			switch strings.ToLower(line) {
			case "p":
				h.selectHandler.SetSelectType(TypeAsset)
				h.selectHandler.Search("")
				continue
			case "b":
				h.selectHandler.MovePrePage()
				continue
			case "n":
				h.selectHandler.MoveNextPage()
				continue
			case "g":
				h.displayNodeTree()
				continue
			case "h":
				h.displayHelp()
				initialed = false
				continue
			case "r":
				h.refreshAssetsAndNodesData()
				continue
			case "q":
				return
			}
		default:
			switch {
			case line == "exit", line == "quit":
				return
			case strings.Index(line, "/") == 0:
				if strings.Index(line[1:], "/") == 0 {
					line = strings.TrimSpace(line[2:])
					h.selectHandler.SearchAgain(line)
					continue
				}
				line = strings.TrimSpace(line[1:])
				h.selectHandler.Search(line)
				continue
			case strings.Index(line, "g") == 0:
				searchWord := strings.TrimSpace(strings.TrimPrefix(line, "g"))
				if num, err := strconv.Atoi(searchWord); err == nil {
					if num > 0 && num <= len(h.nodes) {
						selectedNode := h.nodes[num-1]
						h.selectHandler.SetNode(selectedNode)
						h.selectHandler.Search("")
						continue
					}
				}
			}
		}
		h.selectHandler.SearchOrProxy(line)
	}
}

func (h *InteractiveHandler) checkMaxIdleTime(checkChan <-chan bool) {
	maxIdleMinutes := h.terminalConf.MaxIdleTime
	checkMaxIdleTime(maxIdleMinutes, h.user, h.sess.Sess, checkChan)
}

func (h *InteractiveHandler) displayNodeTree() {
	tree := ConstructNodeTree(h.nodes)
	_, _ = io.WriteString(h.term, "\n\r"+"Node: [ ID.Name(Asset amount) ]")
	_, _ = io.WriteString(h.term, tree.String())
	_, err := io.WriteString(h.term, "Tips: Enter g+NodeID to display the host under the node, such as g1"+"\n\r")
	if err != nil {
		log.Info.Print("displayAssetNodes err:", err)
	}
}
