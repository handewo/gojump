package core

import (
	"fmt"
	"strings"

	"github.com/handewo/gojump/pkg/model"
	"golang.org/x/crypto/bcrypt"
)

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

func (c *Core) QueryAssetUserExpire(userID string, assetID string) (*model.ExpireInfo,
	error) {
	var expireat int64
	err := c.db.QueryOneField(&expireat,
		"SELECT expireat FROM ASSETUSERINFO where userid = ? and assetid = ?", userID, assetID)
	if err != nil {
		return nil, err
	}
	return &model.ExpireInfo{
		ExpireAt: expireat,
	}, nil
}

func (c *Core) GetTerminalConfig() (model.TerminalConfig, error) {
	t := model.TerminalConfig{}
	err := c.db.QueryStruct(&t, "SELECT * FROM TERMINALCONF")
	if err != nil {
		return model.TerminalConfig{}, err
	}
	return t, nil
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

func (c *Core) QueryDirectLoginInfo(userID string, directLogin map[string]string) (asset model.Asset,
	sysUser model.SystemUser, err error) {
	asset, err = c.GetAssetByName(directLogin["asset"])
	if err != nil {
		return
	}
	if asset.ID == "" {
		err = fmt.Errorf("asset %s not found", directLogin["asset"])
		return
	}

	sysUsers, err := c.GetSystemUsersByUserIdAndAssetId(userID, asset.ID)
	if err != nil {
		return
	}
	for _, v := range sysUsers {
		if v.Username == directLogin["sysuser"] {
			sysUser = v
		}
	}
	if sysUser.ID == "" {
		err = fmt.Errorf("system user %s not found", directLogin["sysuser"])
		return
	}
	return
}
