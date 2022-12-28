package initDB

import (
	"fmt"

	"github.com/genjidb/genji"
	"github.com/handewo/gojump/pkg/common"
	"github.com/handewo/gojump/pkg/core"
	"github.com/handewo/gojump/pkg/log"
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
	var err error
	d := model.User{
		ID:       "2",
		Username: "rick",
		Role:     "user",
		ExpireAt: 0,
		OTPLevel: 0,
		IsActive: true,
		NodeIDs:  []string{"1", "2", "10", "59"},
	}
	pass, _ := common.HashPassword("123456")
	us := model.UserSecret{
		UserID:         "2",
		Password:       pass,
		PrivateKey:     "",
		AuthorizedKeys: []string{""},
	}
	a := model.User{
		ID:            "1",
		Username:      "admin",
		Role:          "admin",
		ExpireAt:      0,
		OTPLevel:      0,
		IsActive:      true,
		AddrWhiteList: []string{"127.0.0.1"},
	}
	aus := model.UserSecret{
		UserID:         "1",
		Password:       pass,
		PrivateKey:     "",
		AuthorizedKeys: []string{""},
	}
	f := model.User{
		ID:       "3",
		Username: "jack",
		Role:     "user",
		ExpireAt: 1672216237,
		OTPLevel: 0,
		IsActive: true,
	}
	fs := model.UserSecret{
		UserID:         "3",
		Password:       pass,
		PrivateKey:     "",
		AuthorizedKeys: []string{""},
	}
	err = core.InsertData(db, "INSERT INTO USER VALUES ?", &d)
	if err != nil {
		log.Error.Print(err)
	}
	err = core.InsertData(db, "INSERT INTO USERSECRET VALUES ?", &us)
	if err != nil {
		log.Error.Print(err)
	}
	err = core.InsertData(db, "INSERT INTO USER VALUES ?", &a)
	if err != nil {
		log.Error.Print(err)
	}
	err = core.InsertData(db, "INSERT INTO USERSECRET VALUES ?", &aus)
	if err != nil {
		log.Error.Print(err)
	}
	err = core.InsertData(db, "INSERT INTO USER VALUES ?", &f)
	if err != nil {
		log.Error.Print(err)
	}
	err = core.InsertData(db, "INSERT INTO USERSECRET VALUES ?", &fs)
	if err != nil {
		log.Error.Print(err)
	}
}

func insertNode(db *genji.DB) {
	var err error
	n := model.Node{
		ID:   "1",
		Key:  "1",
		Name: "Default",
	}
	err = core.InsertData(db, "INSERT INTO NODE VALUES ?", &n)
	if err != nil {
		log.Error.Print(err)
	}
}

func insertAssets(db *genji.DB, prefix string, start int, end int) {
	var err error
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
			ID:          fmt.Sprint(i),
			UserID:      "2",
			AssetID:     fmt.Sprint(i),
			ExpireAt:    0,
			SysUserID:   []string{"1", "2"},
			NeedConfirm: true,
		}
		err = core.InsertData(db, "INSERT INTO ASSET VALUES ?", &d)
		if err != nil {
			log.Error.Print(err)
		}
		err = core.InsertData(db, "INSERT INTO ASSETUSERINFO VALUES ?", &ua)
		if err != nil {
			log.Error.Print(err)
		}
		assets = append(assets, fmt.Sprint(i))
	}
	n := model.Node{
		ID:       fmt.Sprint(start),
		Key:      fmt.Sprintf("1:%d", start),
		Name:     prefix,
		AssetIDs: assets,
	}
	err = core.InsertData(db, "INSERT INTO NODE VALUES ?", &n)
	if err != nil {
		log.Error.Print(err)
	}
}

func insertSystemUser(db *genji.DB) {
	var err error
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
	err = core.InsertData(db, "INSERT INTO SYSTEMUSER VALUES ?", &s1)
	if err != nil {
		log.Error.Print(err)
	}
	err = core.InsertData(db, "INSERT INTO SYSTEMUSER VALUES ?", &s2)
	if err != nil {
		log.Error.Print(err)
	}
}
