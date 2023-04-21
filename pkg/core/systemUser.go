package core

import (
	"errors"
	"fmt"

	"github.com/handewo/gojump/pkg/model"
)

func (c *Core) GetSystemUsersByUserIdAndAssetId(userID string, assetID string) (sysUsers []model.SystemUser, err error) {
	sysIDs := make([]string, 0, 5)
	err = c.db.QueryOneField(&sysIDs, "SELECT sysuserid FROM ASSETUSERINFO where userid = ? and assetid = ?", userID, assetID)
	if err != nil {
		return nil, err
	}
	sysUsers, err = c.getSystemUsers(sysIDs)
	return
}

func (c *Core) getSystemUsers(ids []string) ([]model.SystemUser, error) {
	v, err := c.db.QueryStructs(model.SystemUserType, "Select * from SYSTEMUSER WHERE id IN ?", ids)
	if err != nil {
		return nil, err
	}

	sys, ok := v.([]model.SystemUser)
	if !ok {
		return nil, errors.New("invalid value type")
	}
	return sys, err
}

func (c *Core) QueryAllSystemUser() ([]string, error) {
	v, err := c.db.QueryStructs(model.SystemUserType, "SELECT * FROM SYSTEMUSER")
	if err != nil {
		return nil, err
	}

	sys, ok := v.([]model.SystemUser)
	if !ok {
		return nil, errors.New("invalid value type")
	}

	res := make([]string, 0, 10)
	for _, v := range sys {
		pass := ""
		key := ""
		if v.Password != "" {
			pass = "********"
		}
		if v.PrivateKey != "" {
			key = "********"
		}
		s := fmt.Sprintf("%4s|%10s|%8d|%8s|%8s|%11s|%s", v.ID, v.Username,
			v.Priority, v.Protocol, pass, key, v.Comment)
		res = append(res, s)
	}
	return res, nil
}
