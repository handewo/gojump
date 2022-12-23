package server

import (
	"net"

	"github.com/gliderlabs/ssh"
	"github.com/handewo/gojump/pkg/auth"
	"github.com/handewo/gojump/pkg/common"
	"github.com/handewo/gojump/pkg/config"
	"github.com/handewo/gojump/pkg/handler"
	"github.com/handewo/gojump/pkg/log"
	"github.com/handewo/gojump/pkg/model"
)

const (
	nextAuthMethod = "keyboard-interactive"
	ctxID          = "ctxID"
)

func (s *server) GetSSHAddr() string {
	cf := config.GlobalConfig
	return net.JoinHostPort(cf.BindHost, cf.SSHPort)
}

func (s *server) GetSSHSigner() ssh.Signer {
	conf := s.GetTerminalConfig()
	singer, err := ParsePrivateKeyFromString(conf.HostKey)
	if err != nil {
		log.Fatal.Fatal(err)
	}
	return singer
}

func (s *server) PasswordAuth(ctx ssh.Context, password string) bool {
	ctx.SetValue(ctxID, ctx.SessionID())
	tConfig := s.GetTerminalConfig()
	if !tConfig.PasswordAuth {
		log.Info.Print("core disable password auth auth")
		return false
	}
	sshAuthHandler := auth.SSHPasswordAndPublicKeyAuth(s.core)
	return sshAuthHandler(ctx, password, "")
}

func (s *server) PublicKeyAuth(ctx ssh.Context, key ssh.PublicKey) bool {
	ctx.SetValue(ctxID, ctx.SessionID())
	tConfig := s.GetTerminalConfig()
	if !tConfig.PublicKeyAuth {
		log.Info.Print("core disable publickey auth")
		return false
	}
	publicKey := common.Base64Encode(string(key.Marshal()))
	sshAuthHandler := auth.SSHPasswordAndPublicKeyAuth(s.core)
	return sshAuthHandler(ctx, "", publicKey)
}

func (s *server) NextAuthMethodsHandler(ctx ssh.Context) []string {
	return []string{nextAuthMethod}
}

func (s *server) SessionHandler(sess ssh.Session) {
	user, ok := sess.Context().Value(auth.ContextKeyUser).(*model.User)
	if !ok || user.ID == "" {
		log.Error.Printf("SSH User %s not found, exit.", sess.User())
		common.IgnoreErrWriteString(sess, "Not auth user.\n")
		return
	}
	termConf := s.GetTerminalConfig()
	if pty, winChan, isPty := sess.Pty(); isPty {
		interactiveSrv := handler.NewInteractiveHandler(sess, user, s.core, termConf)
		log.Info.Printf("User %s request pty %s", sess.User(), pty.Term)
		go interactiveSrv.WatchWinSizeChange(winChan)
		if user.Username == "admin" {
			interactiveSrv.AdminSystem()
			return
		}
		interactiveSrv.Dispatch()
		common.IgnoreErrWriteWindowTitle(sess, termConf.HeaderTitle)
		return
	} else {
		common.IgnoreErrWriteString(sess, "No PTY requested.\n")
		return
	}

}
