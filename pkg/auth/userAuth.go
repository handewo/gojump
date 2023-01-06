package auth

import (
	"time"

	"github.com/gliderlabs/ssh"
	"github.com/handewo/gojump/pkg/core"
	"github.com/handewo/gojump/pkg/log"
	"github.com/handewo/gojump/pkg/model"
)

const (
	AuthFailed = iota
	AuthBlock
	AuthSuccess
)

type UserAuthClient struct {
	*core.Core
	model.UserClient
}

func (u *UserAuthClient) SetOption(setters ...model.UserClientOption) {
	u.UserClient.SetOption(setters...)
}

func (u *UserAuthClient) Authenticate(ctx ssh.Context) (model.User, int) {
	username := ctx.User()
	if u.Core.UserIsBlocked(username) {
		return model.User{}, AuthBlock
	}
	user, ok := u.Core.UserAuthenticate(username, u.UserClient.Password, u.UserClient.PublicKey)
	if !ok {
		log.Info.Printf("user %s login failed", username)
		return model.User{}, AuthFailed
	}

	if user.AddrWhiteList != nil {
		isMatch := false
		for _, addr := range user.AddrWhiteList {
			if u.UserClient.RemoteAddr == addr {
				isMatch = true
				break
			}
		}
		if !isMatch {
			log.Info.Printf("user %s's IP[%s] is blocked", username, u.UserClient.RemoteAddr)
			return model.User{}, AuthFailed
		}
	}

	t := time.Now().Unix()
	if user.ExpireAt != 0 && user.ExpireAt < t {
		log.Info.Printf("user %s has expired", username)
		return model.User{}, AuthFailed
	}
	return user, AuthSuccess
}
