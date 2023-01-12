package core

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/handewo/gojump/pkg/common"
	"github.com/handewo/gojump/pkg/log"
	"github.com/handewo/gojump/pkg/model"
)

func (c *Core) UserAuthenticate(username string, pass string, pubKey string) (model.User, bool) {
	user, err := c.GetUser(username)
	if err != nil {
		log.Error.Print(err)
		return user, false
	}

	if !user.IsActive {
		return user, false
	}

	switch user.OTPLevel {
	case 1:
		return user, c.verifyOTP(user.Username, pass)
	}
	var sec model.UserSecret
	err = c.db.QueryStruct(&sec, "SELECT * FROM USERSECRET WHERE userid = ?", user.ID)
	if err != nil {
		log.Error.Print(err)
		return user, false
	}
	if sec.UserID == "" {
		log.Error.Printf("querying %s's secret failed", username)
		return user, false
	}
	if pubKey != "" {
		return user, doPublicKeyMatch(sec.AuthorizedKeys, pubKey)
	}
	return user, doPasswordsMatch(sec.Password, pass)
}

func (c *Core) GetUser(name string) (model.User, error) {
	var user model.User
	err := c.db.QueryStruct(&user, "SELECT * FROM USER WHERE username = ?", name)
	if err != nil {
		return user, err
	}
	if user.ID == "" {
		return user, fmt.Errorf("querying user: %s failed", name)
	}
	return user, nil
}

func (c *Core) QueryAllUser() ([]string, error) {
	v, err := c.db.QueryStructs(model.UserType, "SELECT * FROM USER")
	if err != nil {
		return nil, err
	}

	users, ok := v.([]model.User)
	if !ok {
		return nil, errors.New("invalid value type")
	}

	res := make([]string, 0, 10)
	for _, v := range users {
		var ea string
		ea = time.Unix(v.ExpireAt, 0).Format(common.LogFormat)
		if v.ExpireAt == 0 {
			ea = "9999-12-31 23:59:59"
		}
		n := strings.Join(v.NodeIDs, ",")
		l := strings.Join(v.AddrWhiteList, ",")
		s := fmt.Sprintf("%4s|%10s|%6s|%s|%9d|%6v|%16s|%s", v.ID,
			v.Username, v.Role, ea, v.OTPLevel, v.IsActive, n, l)
		res = append(res, s)
	}
	return res, nil
}

func (c *Core) QueryAllUserSecret() ([]string, error) {
	v, err := c.db.QueryStructs(model.UserSecretType, "SELECT * FROM USERSECRET")
	if err != nil {
		return nil, err
	}

	secrets, ok := v.([]model.UserSecret)
	if !ok {
		return nil, errors.New("invalid value type")
	}

	res := make([]string, 0, 10)
	for _, v := range secrets {
		pass := ""
		key := ""
		auth := ""
		if v.Password != "" {
			pass = "********"
		}
		if v.PrivateKey != "" {
			key = "********"
		}
		if v.AuthorizedKeys != nil {
			ks := make([]string, 0, 5)
			for i := 0; i < len(v.AuthorizedKeys); i++ {
				ks = append(ks, "********")
			}
			auth = strings.Join(ks, ",")
		}
		s := fmt.Sprintf("%4s|%8s|%11s|%s", v.UserID,
			pass, key, auth)
		res = append(res, s)
	}
	return res, nil
}
