package model

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const (
	ProtocolSSH = "ssh"
)

type Asset struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Hostname  string   `json:"hostname"`
	IP        string   `json:"ip"`
	Os        string   `json:"os"`
	Comment   string   `json:"comment"`
	Protocols []string `json:"protocols"`
	Platform  string   `json:"platform"`
	IsActive  bool     `json:"is_active"`
}

type AssetUserInfo struct {
	ID          string
	UserID      string
	AssetID     string
	ExpireAt    int64
	SysUserID   []string
	NeedConfirm bool
}

func (a *Asset) String() string {
	return fmt.Sprintf("%s(%s)", a.Hostname, a.IP)
}

func (a *Asset) ProtocolPort(protocol string) int {
	for _, item := range a.Protocols {
		if strings.Contains(strings.ToLower(item), strings.ToLower(protocol)) {
			proAndPort := strings.Split(item, "/")
			if len(proAndPort) == 2 {
				if port, err := strconv.Atoi(proAndPort[1]); err == nil {
					return port
				}
			}
		}
	}
	return 0
}

func (a *Asset) IsSupportProtocol(protocol string) bool {
	for _, item := range a.Protocols {
		if strings.Contains(strings.ToLower(item), strings.ToLower(protocol)) {
			return true
		}
	}
	return false
}

type AssetList []Asset

func (a AssetList) SortBy(byIp bool) AssetList {
	var sortedAssets = make(AssetList, len(a))
	copy(sortedAssets, a)
	if byIp {
		sorter := &assetSorter{
			data:   sortedAssets,
			sortBy: assetSortByIP,
		}
		sort.Sort(sorter)
	} else {
		sorter := &assetSorter{
			data:   sortedAssets,
			sortBy: assetSortByHostName,
		}
		sort.Sort(sorter)
	}
	return sortedAssets
}

type assetSorter struct {
	data   []Asset
	sortBy func(asset1, asset2 *Asset) bool
}

func (s *assetSorter) Len() int {
	return len(s.data)
}

func (s *assetSorter) Swap(i, j int) {
	s.data[i], s.data[j] = s.data[j], s.data[i]
}

func (s *assetSorter) Less(i, j int) bool {
	return s.sortBy(&s.data[i], &s.data[j])
}

func assetSortByIP(asset1, asset2 *Asset) bool {
	iIPs := strings.Split(asset1.IP, ".")
	jIPs := strings.Split(asset2.IP, ".")
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

func assetSortByHostName(asset1, asset2 *Asset) bool {
	return asset1.Hostname < asset2.Hostname
}
