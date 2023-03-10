package auth

import (
	"net"

	"github.com/gliderlabs/ssh"
	"github.com/handewo/gojump/pkg/core"
	"github.com/handewo/gojump/pkg/log"
	"github.com/handewo/gojump/pkg/model"
)

const (
	ContextKeyUser   = "CONTEXT_USER"
	ContextKeyClient = "CONTEXT_CLIENT"
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
		user, res := userAuthClient.Authenticate(ctx)
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
