package initDB

import (
	"fmt"

	"github.com/genjidb/genji"
	"github.com/handewo/gojump/pkg/common"
	"github.com/handewo/gojump/pkg/core"
	"github.com/handewo/gojump/pkg/model"
)

func InsertTestData(db *genji.DB) {
	insertAssets(db, "elastic", 2, 9)
	insertAssets(db, "k8s", 10, 58)
	insertAssets(db, "hadoop", 59, 158)

	insertUser(db)

	insertNode(db)

	insertSystemUser(db)
}

func insertUser(db *genji.DB) {
	d := model.User{
		ID:        "1",
		Username:  "rick",
		Role:      "user",
		ExpiredAt: 0,
		OTPLevel:  0,
		IsActive:  true,
		NodeIDs:   []string{"1", "2", "10", "59"},
	}
	pass, _ := common.HashPassword("123456")
	us := model.UserSecret{
		UserID:         "1",
		Password:       pass,
		PrivateKey:     "",
		AuthorizedKeys: []string{""},
	}
	core.InsertData(db, "INSERT INTO USER VALUES ?", &d)
	core.InsertData(db, "INSERT INTO USERSECRET VALUES ?", &us)
}

func insertNode(db *genji.DB) {
	n := model.Node{
		ID:       "1",
		Key:      "1",
		Name:     "Default",
		AssetIDs: []string{"200"},
	}
	core.InsertData(db, "INSERT INTO NODE VALUES ?", &n)
}

func insertAssets(db *genji.DB, prefix string, start int, end int) {
	assets := make([]string, 0, end-start+1)
	for i := start; i < end; i++ {
		d := model.Asset{
			ID:        fmt.Sprint(i),
			Name:      fmt.Sprintf("%s%d", prefix, i),
			Hostname:  fmt.Sprintf("%s%d", prefix, i),
			IP:        fmt.Sprintf("192.168.2.%d", i),
			Os:        "CentOS 7",
			Comment:   "",
			Protocols: []string{"ssh/22"},
			Platform:  "Linux",
			IsActive:  true,
		}
		ua := model.AssetUserInfo{
			ID:        fmt.Sprint(i),
			UserID:    "1",
			AssetID:   fmt.Sprint(i),
			ExpireAt:  0,
			SysUserID: []string{"1"},
		}
		core.InsertData(db, "INSERT INTO ASSET VALUES ?", &d)
		core.InsertData(db, "INSERT INTO ASSETUSERINFO VALUES ?", &ua)
		assets = append(assets, fmt.Sprint(i))
	}
	n := model.Node{
		ID:       fmt.Sprint(start),
		Key:      fmt.Sprintf("1:%d", start),
		Name:     prefix,
		AssetIDs: assets,
	}
	core.InsertData(db, "INSERT INTO NODE VALUES ?", &n)
}

func insertSystemUser(db *genji.DB) {
	s1 := model.SystemUser{
		ID:         "1",
		Username:   "root",
		Priority:   1,
		Protocol:   "ssh",
		Comment:    "",
		Password:   "",
		PrivateKey: "",
	}
	s2 := model.SystemUser{
		ID:         "2",
		Username:   "rick",
		Priority:   1,
		Protocol:   "ssh",
		Comment:    "",
		Password:   "",
		PrivateKey: "",
	}
	core.InsertData(db, "INSERT INTO SYSTEMUSER VALUES ?", &s1)
	core.InsertData(db, "INSERT INTO SYSTEMUSER VALUES ?", &s2)
}
