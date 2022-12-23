package core

import (
	"fmt"

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
	err = queryStructFromDb(c.db, &sec, "SELECT * FROM USERSECRET WHERE userid = ?", user.ID)
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
	err := queryStructFromDb(c.db, &user, "SELECT * FROM USER WHERE username = ?", name)
	if err != nil {
		return user, err
	}
	if user.ID == "" {
		return user, fmt.Errorf("querying user: %s failed", name)
	}
	return user, nil
}
