package core

import (
	"strings"

	"github.com/genjidb/genji"
	"github.com/genjidb/genji/document"
	"github.com/genjidb/genji/types"
	"github.com/handewo/gojump/pkg/model"
	"golang.org/x/crypto/bcrypt"
)

type Model interface {
	model.Asset | model.Node | model.User | model.SystemUser | model.AssetUserInfo
}

func queryOneFieldFromDb(db *genji.DB, f interface{}, sql string, cond ...interface{}) error {
	res, err := db.Query(sql, cond...)
	if err != nil {
		return err
	}
	defer res.Close()
	err = res.Iterate(func(d types.Document) error {
		err = document.Scan(d, f)
		return err
	})
	return err
}

func queryStructFromDb(db *genji.DB, st interface{}, sql string, cond ...interface{}) error {
	res, err := db.Query(sql, cond...)
	if err != nil {
		return err
	}
	defer res.Close()
	err = res.Iterate(func(d types.Document) error {
		err = document.StructScan(d, st)
		return err
	})
	return err
}

func queryStructsFromDb[M Model](db *genji.DB, sql string, cond ...interface{}) ([]M, error) {
	res, err := db.Query(sql, cond...)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	models := make([]M, 0, 10)
	err = res.Iterate(func(d types.Document) error {
		var m M
		err = document.StructScan(d, &m)
		if err != nil {
			return err
		}
		models = append(models, m)
		return nil
	})
	return models, err
}

func InsertData(db *genji.DB, sql string, data ...interface{}) error {
	err := db.Exec(sql, data...)
	return err
}

func CreateSchema(db *genji.DB, sql string) error {
	err := db.Exec(sql)
	return err
}

// Check if two passwords match using Bcrypt's CompareHashAndPassword
// which return nil on success and an error on failure.
func doPasswordsMatch(hashedPassword, currPassword string) bool {
	if currPassword == "" {
		return false
	}
	err := bcrypt.CompareHashAndPassword(
		[]byte(hashedPassword), []byte(currPassword))
	return err == nil
}

func doPublicKeyMatch(authKeys []string, currPubkey string) bool {
	for _, v := range authKeys {
		keyInfo := strings.Split(v, " ")
		if len(keyInfo) < 2 {
			continue
		}
		if keyInfo[1] == currPubkey {
			return true
		}
	}
	return false
}

func (c *Core) ValidateAssetConnectPermission(userID string, assetID string) (*model.ExpireInfo,
	error) {
	var expireat int64

	err := queryOneFieldFromDb(c.db, &expireat,
		"SELECT expireat FROM ASSETUSERINFO where userid = ? and assetid = ?", userID, assetID)
	if err != nil {
		return nil, err
	}
	return &model.ExpireInfo{
		ExpireAt: expireat,
	}, nil
}
