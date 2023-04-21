package auth

import (
	"net"
	"strings"

	"github.com/gliderlabs/ssh"
	"github.com/handewo/gojump/pkg/core"
	"github.com/handewo/gojump/pkg/log"
	"github.com/handewo/gojump/pkg/model"
)

const (
	ContextKeyUser              = "CONTEXT_USER"
	ContextKeyClient            = "CONTEXT_CLIENT"
	ContextKeyDirectLoginFormat = "CONTEXT_DIRECT_LOGIN_FORMAT"
)

type SSHAuthFunc func(ctx ssh.Context, password, publicKey string) bool

func SSHPasswordAndPublicKeyAuth(c *core.Core) SSHAuthFunc {
	return func(ctx ssh.Context, password, publicKey string) bool {
		remoteAddr, _, _ := net.SplitHostPort(ctx.RemoteAddr().String())
		username := ctx.User()
		authMethod := "publickey"
		if password != "" {
			authMethod = "password"
		}
		if res, ok := parseUserFormatBySeparator(ctx.User()); ok {
			ctx.SetValue(ContextKeyDirectLoginFormat, res)
			username = res["username"]
		}
		userAuthClient, ok := ctx.Value(ContextKeyClient).(*UserAuthClient)
		if !ok {
			userAuthClient = &UserAuthClient{
				Core:       c,
				UserClient: model.UserClient{},
			}
			ctx.SetValue(ContextKeyClient, userAuthClient)
		}
		userAuthClient.SetOption(model.UserClientPassword(password),
			model.UserClientPublicKey(publicKey), model.UserClientRemoteAddr(remoteAddr))
		user, res := userAuthClient.Authenticate(username)
		switch res {
		case AuthSuccess:
			ctx.SetValue(ContextKeyUser, &user)
			c.AuthenticationLog(username, authMethod, remoteAddr)
			log.Info.Printf("SSH conn[%s] %s for %s from %s", ctx.SessionID()[:10],
				authMethod, username, remoteAddr)
			return true
		case AuthFailed:
			log.Info.Printf("SSH conn[%s] %s for %s from %s", ctx.SessionID()[:10],
				authMethod, username, remoteAddr)
			c.LimitTryLogin(username)
		case AuthBlock:
		}
		return false
	}
}

func parseUserFormatBySeparator(s string) (map[string]string, bool) {
	authInfos := strings.Split(s, "@")
	if len(authInfos) != 3 {
		return nil, false
	}
	res := map[string]string{
		"username": authInfos[0],
		"sysuser":  authInfos[1],
		"asset":    authInfos[2],
	}
	return res, true
}
