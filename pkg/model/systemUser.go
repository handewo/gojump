package model

import (
	"fmt"
	"sort"
	"strings"
)

const LoginModeManual = "manual"

const (
	AllAction      = "all"
	ConnectAction  = "connect"
	UploadAction   = "upload_file"
	DownloadAction = "download_file"
)

type SystemUser struct {
	ID         string `json:"id"`
	Username   string `json:"username"`
	Priority   int    `json:"priority"`
	Protocol   string `json:"protocol"`
	Comment    string `json:"comment"`
	Password   string `json:"-"`
	PrivateKey string `json:"-"`
}

func (s *SystemUser) String() string {
	return fmt.Sprint(s.Username)
}

func (s *SystemUser) IsProtocol(p string) bool {
	return strings.EqualFold(s.Protocol, p)
}

type systemUserSortBy func(sys1, sys2 *SystemUser) bool

func (by systemUserSortBy) Sort(sysUsers []SystemUser) {
	nodeSorter := &systemUserSorter{
		users:  sysUsers,
		sortBy: by,
	}
	sort.Sort(nodeSorter)
}

type systemUserSorter struct {
	users  []SystemUser
	sortBy systemUserSortBy
}

func (s *systemUserSorter) Len() int {
	return len(s.users)
}

func (s *systemUserSorter) Swap(i, j int) {
	s.users[i], s.users[j] = s.users[j], s.users[i]
}

func (s *systemUserSorter) Less(i, j int) bool {
	return s.sortBy(&s.users[i], &s.users[j])
}

func systemUserPrioritySort(use1, user2 *SystemUser) bool {
	return use1.Priority < user2.Priority
}

func SortSystemUserByPriority(sysUsers []SystemUser) {
	systemUserSortBy(systemUserPrioritySort).Sort(sysUsers)
}
