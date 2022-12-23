package handler

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/handewo/gojump/pkg/common"
	"github.com/handewo/gojump/pkg/log"
	"github.com/handewo/gojump/pkg/model"
	"github.com/handewo/gojump/pkg/proxy"
)

func (u *UserSelectHandler) displayNodeAssetResult(searchHeader string) {
	term := u.h.term
	if len(u.currentResult) == 0 {
		noNodeAssets := fmt.Sprintf("%s node has no assets", u.selectedNode.Name)
		common.IgnoreErrWriteString(term, common.WrapperString(noNodeAssets, common.Red))
		common.IgnoreErrWriteString(term, common.CharNewLine)
		common.IgnoreErrWriteString(term, common.WrapperString(searchHeader, common.Green))
		common.IgnoreErrWriteString(term, common.CharNewLine)
		return
	}
	u.displaySortedAssets(searchHeader)
}

func (u *UserSelectHandler) searchLocalAsset(searches ...string) []map[string]interface{} {
	fields := map[string]struct{}{
		"name":     {},
		"hostname": {},
		"ip":       {},
		"platform": {},
		"comment":  {},
	}
	return u.searchLocalFromFields(fields, searches...)
}

func (u *UserSelectHandler) displayAssetResult(searchHeader string) {
	term := u.h.term
	if len(u.currentResult) == 0 {
		noAssets := ("No Assets")
		common.IgnoreErrWriteString(term, common.WrapperString(noAssets, common.Red))
		common.IgnoreErrWriteString(term, common.CharNewLine)
		common.IgnoreErrWriteString(term, common.WrapperString(searchHeader, common.Green))
		common.IgnoreErrWriteString(term, common.CharNewLine)
		return
	}
	u.displaySortedAssets(searchHeader)
}

func (u *UserSelectHandler) displaySortedAssets(searchHeader string) {
	assetListSortByIp := u.h.terminalConf.AssetListSortByIp
	if assetListSortByIp {
		sortedAsset := IPAssetList(u.currentResult)
		sort.Sort(sortedAsset)
		u.currentResult = sortedAsset
	} else {
		sortedAsset := HostnameAssetList(u.currentResult)
		sort.Sort(sortedAsset)
		u.currentResult = sortedAsset
	}
	term := u.h.term
	currentPage := u.CurrentPage()
	totalCount := u.TotalCount()
	totalPage := u.TotalPage()

	idLabel := ("ID")
	hostLabel := ("Hostname")
	ipLabel := ("IP")
	platformLabel := ("Platform")
	commentLabel := ("Comment")

	Labels := []string{idLabel, hostLabel, ipLabel, platformLabel, commentLabel}
	fields := []string{"ID", "Hostname", "IP", "Platform", "Comment"}
	data := make([]map[string]string, len(u.currentResult))
	for i, j := range u.currentResult {
		row := make(map[string]string)
		row["ID"] = strconv.Itoa(i + 1)
		fieldMap := map[string]string{
			"hostname": "Hostname",
			"ip":       "IP",
			"platform": "Platform",
			"comment":  "Comment",
		}
		row = convertMapItemToRow(j, fieldMap, row)
		row["Comment"] = joinMultiLineString(row["Comment"])
		data[i] = row
	}
	w, _ := term.GetSize()
	caption := fmt.Sprintf(("Page: %d, Total Page: %d, Total Count: %d"),
		currentPage, totalPage, totalCount)

	caption = common.WrapperString(caption, common.Green)
	table := common.WrapperTable{
		Fields: fields,
		Labels: Labels,
		FieldsSize: map[string][3]int{
			"ID":       {0, 0, 5},
			"Hostname": {0, 40, 0},
			"IP":       {0, 8, 40},
			"Platform": {0, 8, 0},
			"Comment":  {0, 0, 0},
		},
		Data:        data,
		TotalSize:   w,
		Caption:     caption,
		TruncPolicy: common.TruncMiddle,
	}
	table.Initial()
	loginTip := ("Enter ID number directly login the asset, multiple search use // + field, such as: //16")
	pageActionTip := ("Page up: b	Page down: n")

	_, _ = term.Write([]byte(common.CharClear))
	_, _ = term.Write([]byte(table.Display()))
	common.IgnoreErrWriteString(term, common.WrapperString(loginTip, common.Green))
	common.IgnoreErrWriteString(term, common.CharNewLine)
	common.IgnoreErrWriteString(term, common.WrapperString(pageActionTip, common.Green))
	common.IgnoreErrWriteString(term, common.CharNewLine)
	common.IgnoreErrWriteString(term, common.WrapperString(searchHeader, common.Green))
	common.IgnoreErrWriteString(term, common.CharNewLine)
}

func (u *UserSelectHandler) proxyAsset(asset model.Asset) {
	systemUsers, err := u.h.core.GetSystemUsersByUserIdAndAssetId(u.user.ID, asset.ID)
	if err != nil {
		return
	}
	selectedSystemUser, ok := u.h.chooseSystemUser(systemUsers)

	if !ok {
		return
	}
	srv, err := proxy.NewServer(u.h.sess,
		u.h.core,
		proxy.ConnectUser(u.h.user),
		proxy.ConnectAsset(&asset),
		proxy.ConnectSystemUser(&selectedSystemUser),
	)
	if err != nil {
		log.Error.Print(err)
		return
	}
	srv.Proxy()
	log.Info.Printf("Request %s: asset %s proxy end", u.h.sess.Uuid[:8], asset.Hostname)

}

var (
	_ sort.Interface = (HostnameAssetList)(nil)
	_ sort.Interface = (IPAssetList)(nil)
)

type HostnameAssetList []map[string]interface{}

func (l HostnameAssetList) Len() int {
	return len(l)
}

func (l HostnameAssetList) Less(i, j int) bool {
	iHostnameValue := l[i]["hostname"]
	jHostnameValue := l[j]["hostname"]
	iHostname, ok := iHostnameValue.(string)
	if !ok {
		return false
	}
	jHostname, ok := jHostnameValue.(string)
	if !ok {
		return false
	}
	return CompareString(iHostname, jHostname)
}

func (l HostnameAssetList) Swap(i, j int) {
	l[j], l[i] = l[i], l[j]
}

type IPAssetList []map[string]interface{}

func (l IPAssetList) Len() int {
	return len(l)
}

func (l IPAssetList) Less(i, j int) bool {
	iIPValue := l[i]["ip"]
	jIPValue := l[j]["ip"]
	iIP, ok := iIPValue.(string)
	if !ok {
		return false
	}
	jIP, ok := jIPValue.(string)
	if !ok {
		return false
	}
	return CompareIP(iIP, jIP)
}

func (l IPAssetList) Swap(i, j int) {
	l[j], l[i] = l[i], l[j]
}

func CompareIP(ipA, ipB string) bool {
	iIPs := strings.Split(ipA, ".")
	jIPs := strings.Split(ipB, ".")
	for i := 0; i < len(iIPs); i++ {
		if i >= len(jIPs) {
			return false
		}
		if len(iIPs[i]) == len(jIPs[i]) {
			if iIPs[i] == jIPs[i] {
				continue
			} else {
				return iIPs[i] < jIPs[i]
			}
		} else {
			return len(iIPs[i]) < len(jIPs[i])
		}

	}
	return true
}

func CompareString(a, b string) bool {
	return a < b
}
