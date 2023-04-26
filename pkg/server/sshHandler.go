package server

import (
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/gliderlabs/ssh"
	"github.com/handewo/gojump/pkg/auth"
	"github.com/handewo/gojump/pkg/common"
	"github.com/handewo/gojump/pkg/config"
	"github.com/handewo/gojump/pkg/handler"
	"github.com/handewo/gojump/pkg/log"
	"github.com/handewo/gojump/pkg/model"
	"github.com/handewo/gojump/pkg/srvconn"
	gossh "golang.org/x/crypto/ssh"
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
	directLogin := sess.Context().Value(auth.ContextKeyDirectLoginFormat)

	if pty, winChan, isPty := sess.Pty(); isPty {
		if directLogin, ok := directLogin.(map[string]string); ok {
			handler, err := handler.NewDirectHandler(sess, s.core, user, directLogin)
			if err != nil {
				log.Error.Printf("User %s direct handler err: %s", user.Username, err)
				return
			}
			if err := handler.Proxy(); err != nil {
				log.Error.Printf("User %s direct proxy err: %s", user.Username, err)
			}
			return
		}
		interactiveSrv := handler.NewInteractiveHandler(sess, user, s.core, termConf)
		defer s.core.InteractiveLog(sess.User())
		log.Debug.Printf("User %s request pty %s", sess.User(), pty.Term)
		go interactiveSrv.WatchWinSizeChange(winChan)
		if user.Username == "admin" {
			interactiveSrv.AdminSystem()
			return
		}
		interactiveSrv.Dispatch()
		common.IgnoreErrWriteWindowTitle(sess, termConf.HeaderTitle)
		return
	}

	if !config.GlobalConfig.EnableLocalPortForward {
		common.IgnoreErrWriteString(sess, "No PTY requested.\n")
		return
	}

	if directLogin, ok := directLogin.(map[string]string); ok {
		if err := s.proxyVscode(sess, user, directLogin); err != nil {
			log.Error.Printf("User %s login vscode err: %s", user.Username, err)
		}
		return
	}

}

func (s *server) LocalPortForwardingPermission(ctx ssh.Context, destinationHost string, destinationPort uint32) bool {
	return config.GlobalConfig.EnableLocalPortForward
}

func (s *server) DirectTCPIPChannelHandler(ctx ssh.Context, newChan gossh.NewChannel, destAddr string) {
	reqId, ok := ctx.Value(ctxID).(string)
	if !ok {
		_ = newChan.Reject(gossh.Prohibited, "port forwarding is disabled")
		return
	}
	vsReq := s.getVSCodeReq(reqId)
	if vsReq == nil {
		_ = newChan.Reject(gossh.Prohibited, "port forwarding is disabled")
		return
	}
	dConn, err := vsReq.client.Dial("tcp", destAddr)
	if err != nil {
		_ = newChan.Reject(gossh.ConnectionFailed, err.Error())
		return
	}
	defer dConn.Close()
	ch, reqs, err := newChan.Accept()
	if err != nil {
		_ = dConn.Close()
		_ = newChan.Reject(gossh.ConnectionFailed, err.Error())
		return
	}
	log.Info.Printf("User %s start port forwarding from (%s) to (%s)", vsReq.user,
		vsReq.client, destAddr)
	defer ch.Close()
	go gossh.DiscardRequests(reqs)
	go func() {
		defer ch.Close()
		defer dConn.Close()
		_, _ = io.Copy(ch, dConn)
	}()
	_, _ = io.Copy(dConn, ch)
	log.Info.Printf("User %s end port forwarding from (%s) to (%s)", vsReq.user,
		vsReq.client, destAddr)
}

func (s *server) proxyVscode(sess ssh.Session, user *model.User, directLogin map[string]string) error {
	asset, sysUser, err := s.core.QueryDirectLoginInfo(user.ID, directLogin)
	if err != nil {
		return err
	}
	ctxId, ok := sess.Context().Value(ctxID).(string)
	if !ok {
		return errors.New("not found ctxID")
	}
	expireInfo, err := s.core.QueryAssetUserExpire(user.ID, asset.ID)
	if err != nil {
		return fmt.Errorf("get asset expire info err: %s", err)
	}
	vscodePerm, err := s.core.QueryAssetUserVscodePerm(user.ID, asset.ID)
	if err != nil {
		return fmt.Errorf("get asset expire info err: %s", err)
	}
	if !vscodePerm {
		log.Info.Printf("%s has no permission to login to %s in vscode", user.Username, asset.Name)
		return nil
	}
	sshAuthOpts := srvconn.BuildSSHClientOptions(&asset, &sysUser)
	sshClient, err := srvconn.NewSSHClient(sshAuthOpts...)
	if err != nil {
		return fmt.Errorf("get SSH Client failed: %s", err)
	}
	defer sshClient.Close()
	vsReq := &vscodeReq{
		reqId:      ctxId,
		user:       user,
		client:     sshClient,
		expireInfo: expireInfo,
	}
	return s.proxyVscodeShell(sess, vsReq, sshClient)
}

func (s *server) proxyVscodeShell(sess ssh.Session, vsReq *vscodeReq, sshClient *srvconn.SSHClient) error {
	goSess, err := sshClient.AcquireSession()
	if err != nil {
		return fmt.Errorf("get SSH session failed: %s", err)
	}
	defer goSess.Close()
	defer sshClient.ReleaseSession(goSess)
	stdOut, err := goSess.StdoutPipe()
	if err != nil {
		return fmt.Errorf("get SSH session StdoutPipe failed: %s", err)
	}
	stdin, err := goSess.StdinPipe()
	if err != nil {
		return fmt.Errorf("get SSH session StdinPipe failed: %s", err)
	}
	err = goSess.Shell()
	if err != nil {
		return fmt.Errorf("get SSH session shell failed: %s", err)
	}
	log.Info.Printf("User %s start vscode request to %s", vsReq.user, sshClient)

	s.addVSCodeReq(vsReq)
	defer s.deleteVSCodeReq(vsReq)
	go func() {
		_, _ = io.Copy(stdin, sess)
		log.Info.Printf("User %s vscode request %s stdin end", vsReq.user, sshClient)
	}()
	go func() {
		_, _ = io.Copy(sess, stdOut)
		log.Info.Printf("User %s vscode request %s stdOut end", vsReq.user, sshClient)
	}()
	maxConnectTime := time.Duration(8) * time.Hour
	timeout := time.NewTicker(maxConnectTime)
	defer timeout.Stop()
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-sess.Context().Done():
			log.Info.Printf("SSH conn[%s] User %s end vscode request %s as session done",
				vsReq.reqId[:10], vsReq.user, sshClient)
			return nil
		case <-timeout.C:
			log.Info.Printf("SSH conn[%s] User %s end vscode request %s as timeout",
				vsReq.reqId[:10], vsReq.user, sshClient)
			return nil
		case now := <-ticker.C:
			if vsReq.expireInfo.IsExpired(now) {
				log.Info.Printf("SSH conn[%s] User %s end vscode request %s as permission has expired",
					vsReq.reqId[:10], vsReq.user, sshClient)
				return nil
			}
		}
	}
}
