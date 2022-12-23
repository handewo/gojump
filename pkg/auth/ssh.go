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
	return func(ctx ssh.Context, password, publicKey string) (res bool) {
		remoteAddr, _, _ := net.SplitHostPort(ctx.RemoteAddr().String())
		res = false
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
		log.Info.Printf("SSH conn[%s] authenticating user %s %s", ctx.SessionID()[:10], username, authMethod)
		user, res := userAuthClient.Authenticate(ctx)
		log.Info.Printf("SSH conn[%s] %s for %s from %s", ctx.SessionID()[:10],
			authMethod, username, remoteAddr)
		if res {
			ctx.SetValue(ContextKeyUser, &user)
			return
		}
		return
	}
}
