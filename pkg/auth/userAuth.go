package auth

import (
	"time"

	"github.com/gliderlabs/ssh"
	"github.com/handewo/gojump/pkg/core"
	"github.com/handewo/gojump/pkg/log"
	"github.com/handewo/gojump/pkg/model"
)

type UserAuthClient struct {
	*core.Core
	model.UserClient
}

func (u *UserAuthClient) SetOption(setters ...model.UserClientOption) {
	u.UserClient.SetOption(setters...)
}

func (u *UserAuthClient) Authenticate(ctx ssh.Context) (user model.User, res bool) {
	username := ctx.User()
	user, res = u.Core.UserAuthenticate(username, u.UserClient.Password, u.UserClient.PublicKey)
	if !res {
		log.Info.Printf("user %s login failed", username)
		return model.User{}, false
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
			return model.User{}, false
		}
	}

	t := time.Now().Unix()
	if user.ExpiredAt != 0 && user.ExpiredAt < t {
		log.Info.Printf("user %s has expired", username)
		return model.User{}, false
	}
	return user, res
}
