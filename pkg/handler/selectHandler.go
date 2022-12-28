package handler

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/handewo/gojump/pkg/common"
	"github.com/handewo/gojump/pkg/log"
	"github.com/handewo/gojump/pkg/model"
)

type selectType int

const (
	TypeAsset selectType = iota + 1
	TypeNodeAsset
)

type UserSelectHandler struct {
	user *model.User
	h    *InteractiveHandler

	currentType selectType
	searchKeys  []string

	hasPre  bool
	hasNext bool

	allLocalData []map[string]interface{}

	selectedNode  model.Node
	currentResult []map[string]interface{}

	*pageInfo
}

func (u *UserSelectHandler) SetSelectType(s selectType) {
	u.currentType = s
	switch s {
	case TypeAsset:
		u.AutoCompletion()
		u.h.term.SetPrompt("[Host]> ")
	case TypeNodeAsset:
		u.h.term.SetPrompt("[Host]> ")
	}
}

func (u *UserSelectHandler) AutoCompletion() {
	assets := u.Retrieve(0, 0, "")
	suggests := make([]string, 0, len(assets))

	for _, v := range assets {
		switch u.currentType {
		case TypeAsset, TypeNodeAsset:
			suggests = append(suggests, v["hostname"].(string))
		default:
			suggests = append(suggests, v["name"].(string))
		}
	}

	sort.Strings(suggests)
	u.h.term.AutoCompleteCallback = func(line string, pos int, key rune) (newLine string, newPos int, ok bool) {
		if key == 9 {
			termWidth, _ := u.h.term.GetSize()
			if len(line) >= 1 {
				sugs := common.FilterPrefix(suggests, line)
				if len(sugs) >= 1 {
					commonPrefix := common.LongestCommonPrefix(sugs)
					switch u.currentType {
					case TypeAsset, TypeNodeAsset:
						fmt.Fprintf(u.h.term, "%s%s\n%s\n", "[Host]> ", line, common.Pretty(sugs, termWidth))
					}
					return commonPrefix, len(commonPrefix), true
				}
			}
		}

		return newLine, newPos, false
	}
}

func (u *UserSelectHandler) SetNode(node model.Node) {
	u.SetSelectType(TypeNodeAsset)
	u.selectedNode = node
}

func (u *UserSelectHandler) HasJustOneAsset() bool {
	return len(u.allLocalData) == 1
}

func (u *UserSelectHandler) SetAllLocalData(data []map[string]interface{}) {
	// 使用副本
	u.allLocalData = make([]map[string]interface{}, len(data))
	copy(u.allLocalData, data)
}

func (u *UserSelectHandler) MoveNextPage() {
	if u.HasNext() {
		offset := u.CurrentOffSet()
		newPageSize := getPageSize(u.h.term, u.h.terminalConf)
		u.currentResult = u.Retrieve(newPageSize, offset, u.searchKeys...)
	}
	u.DisplayCurrentResult()
}

func (u *UserSelectHandler) MovePrePage() {
	if u.HasPrev() {
		offset := u.CurrentOffSet()
		newPageSize := getPageSize(u.h.term, u.h.terminalConf)
		start := offset - newPageSize*2
		if start <= 0 {
			start = 0
		}
		u.currentResult = u.Retrieve(newPageSize, start, u.searchKeys...)
	}
	u.DisplayCurrentResult()
}

func (u *UserSelectHandler) Search(key string) {
	newPageSize := getPageSize(u.h.term, u.h.terminalConf)
	u.currentResult = u.Retrieve(newPageSize, 0, key)
	u.searchKeys = []string{key}
	u.DisplayCurrentResult()
}

func (u *UserSelectHandler) SearchAgain(key string) {
	u.searchKeys = append(u.searchKeys, key)
	newPageSize := getPageSize(u.h.term, u.h.terminalConf)
	u.currentResult = u.Retrieve(newPageSize, 0, u.searchKeys...)
	u.DisplayCurrentResult()
}

func (u *UserSelectHandler) SearchOrProxy(key string) {
	if indexNum, err := strconv.Atoi(key); err == nil && len(u.currentResult) > 0 {
		if indexNum > 0 && indexNum <= len(u.currentResult) {
			u.Proxy(u.currentResult[indexNum-1])
			return
		}
	}

	newPageSize := getPageSize(u.h.term, u.h.terminalConf)
	currentResult := u.Retrieve(newPageSize, 0, key)
	u.currentResult = currentResult
	u.searchKeys = []string{key}
	if len(currentResult) == 1 {
		u.Proxy(currentResult[0])
		return
	}

	switch u.currentType {
	case TypeAsset:
		if strings.TrimSpace(key) != "" {
			if ret, ok := getUniqueAssetFromKey(key, currentResult); ok {
				u.Proxy(ret)
				return
			}
		}
	}
	u.DisplayCurrentResult()
}

func (u *UserSelectHandler) HasPrev() bool {
	return u.hasPre
}

func (u *UserSelectHandler) HasNext() bool {
	return u.hasNext
}

func (u *UserSelectHandler) DisplayCurrentResult() {
	searchHeader := fmt.Sprintf("Search: %s", strings.Join(u.searchKeys, " "))
	switch u.currentType {
	case TypeNodeAsset:
		u.displayNodeAssetResult(searchHeader)
	case TypeAsset:
		u.displayAssetResult(searchHeader)
	default:
		log.Error.Print("Display unknown type")
	}
}

func (u *UserSelectHandler) Proxy(target map[string]interface{}) {
	targetId := target["id"].(string)
	switch u.currentType {
	case TypeAsset, TypeNodeAsset:
		asset, err := u.h.core.GetAssetById(targetId)
		if err != nil || asset.ID == "" {
			log.Error.Printf("Select asset %s not found", targetId)
			return
		}
		if !asset.IsActive {
			log.Debug.Printf("Select asset %s is inactive", targetId)
			msg := "The asset is inactive"
			_, _ = u.h.term.Write([]byte(msg))
			return
		}
		msg := fmt.Sprintf("Login to %s%s", asset.Name, common.CharNewLine)
		_, _ = u.h.term.Write([]byte(msg))
		u.proxyAsset(asset)
	default:
		log.Error.Printf("Select unknown type for target id %s", targetId)
	}
}

func (u *UserSelectHandler) Retrieve(pageSize, offset int, searches ...string) []map[string]interface{} {
	if pageSize <= 0 {
		pageSize = PAGESIZEALL
	}
	if offset < 0 {
		offset = 0
	}

	searchResult := u.retrieveLocal(searches...)
	var (
		totalData       []map[string]interface{}
		currentTotal    int
		currentOffset   int
		currentPageSize int
		totalCount      int
	)
	totalCount = len(searchResult)

	if offset < len(searchResult) {
		totalData = searchResult[offset:]
	}
	currentPageSize = pageSize
	currentData := totalData
	currentTotal = len(totalData)

	if currentPageSize < 0 || currentPageSize == PAGESIZEALL {
		currentPageSize = len(totalData)
	}
	if currentTotal > currentPageSize {
		currentData = totalData[:currentPageSize]
	}
	currentOffset = offset + len(currentData)
	u.updatePageInfo(currentPageSize, currentTotal, currentOffset, totalCount)
	u.hasPre = false
	u.hasNext = false
	if u.currentPage > 1 {
		u.hasPre = true
	}
	if u.currentPage < u.totalPage {
		u.hasNext = true
	}
	//log.Debug.Printf("pageSize: %d, total: %d, u.totalPage: %d, u.currentPage: %d, offset: %d",
	// currentPageSize, currentTotal, u.totalPage, u.currentPage, offset)
	return currentData
}

func (u *UserSelectHandler) retrieveLocal(searches ...string) []map[string]interface{} {
	switch u.currentType {
	case TypeAsset:
		return u.searchLocalAsset(searches...)
	case TypeNodeAsset:
		return u.retrieveLocalNodeAsset()
	default:
		// TypeAsset
		u.SetSelectType(TypeAsset)
		log.Debug.Print("Retrieve default local data type: Asset")
		return u.searchLocalAsset(searches...)
	}
}

func (u *UserSelectHandler) searchLocalFromFields(fields map[string]struct{}, searches ...string) []map[string]interface{} {
	items := make([]map[string]interface{}, 0, len(u.allLocalData))
	for _, v := range u.allLocalData {
		if containKeysInMapItemFields(v, fields, searches...) {
			items = append(items, v)
		}
	}
	return items
}

func containKeysInMapItemFields(item map[string]interface{},
	searchFields map[string]struct{}, matchedKeys ...string) bool {

	if len(matchedKeys) == 0 {
		return true
	}
	if len(matchedKeys) == 1 && matchedKeys[0] == "" {
		return true
	}

	for key, value := range item {
		if _, ok := searchFields[key]; ok {
			switch result := value.(type) {
			case string:
				for _, v := range matchedKeys {
					if strings.Contains(result, v) {
						return true
					}
				}
			}
		}
	}
	return false
}

func convertMapItemToRow(item map[string]interface{}, fields map[string]string, row map[string]string) map[string]string {
	for key, value := range item {
		if rowKey, ok := fields[key]; ok {
			switch ret := value.(type) {
			case string:
				row[rowKey] = ret
			case int:
				row[rowKey] = strconv.Itoa(ret)
			}
			continue
		}
	}
	return row
}

func (u *UserSelectHandler) retrieveLocalNodeAsset() []map[string]interface{} {
	ids := u.selectedNode.AssetIDs
	items := make([]map[string]interface{}, 0, 10)
	for _, v := range u.allLocalData {
		if containAssetIDinMap(v, ids) {
			items = append(items, v)
		}
	}
	return items
}

func containAssetIDinMap(item map[string]interface{}, iDs []string) bool {
	id, ok := item["id"]
	if !ok {
		return false
	}
	for _, v := range iDs {
		if id == v {
			return true
		}
	}
	return false
}

func joinMultiLineString(lines string) string {
	lines = strings.ReplaceAll(lines, "\r", "\n")
	lines = strings.ReplaceAll(lines, "\n\n", "\n")
	lineArray := strings.Split(strings.TrimSpace(lines), "\n")
	lineSlice := make([]string, 0, len(lineArray))
	for _, item := range lineArray {
		cleanLine := strings.TrimSpace(item)
		if cleanLine == "" {
			continue
		}
		lineSlice = append(lineSlice, strings.ReplaceAll(cleanLine, " ", ","))
	}
	return strings.Join(lineSlice, "|")
}

func getUniqueAssetFromKey(key string, currentResult []map[string]interface{}) (data map[string]interface{}, ok bool) {
	result := make([]int, 0, len(currentResult))
	for i := range currentResult {
		ip := currentResult[i]["ip"].(string)
		hostname := currentResult[i]["hostname"].(string)
		switch key {
		case ip, hostname:
			result = append(result, i)
		}
	}
	if len(result) == 1 {
		return currentResult[result[0]], true
	}
	return nil, false
}
