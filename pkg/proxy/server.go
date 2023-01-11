package proxy

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync/atomic"
	"time"

	"github.com/handewo/gojump/pkg/auth"
	"github.com/handewo/gojump/pkg/common"
	"github.com/handewo/gojump/pkg/config"
	"github.com/handewo/gojump/pkg/core"
	"github.com/handewo/gojump/pkg/log"
	"github.com/handewo/gojump/pkg/model"
	"github.com/handewo/gojump/pkg/srvconn"
	gossh "golang.org/x/crypto/ssh"
)

var (
	ErrUnMatchProtocol = errors.New("the protocols are not matched")
	ErrPermission      = errors.New("no permission")
	ErrNoAuthInfo      = errors.New("no auth info")
	ErrInvalidPort     = errors.New("invalid port")
)

func NewServer(conn UserConnection, core *core.Core, opts ...ConnectionOption) (*Server, error) {
	connOpts := &ConnectionOptions{}
	for _, setter := range opts {
		setter(connOpts)
	}

	if err := srvconn.IsSupportedProtocol(connOpts.systemUser.Protocol); err != nil {
		log.Error.Printf("Conn[%s] checking protocol %s failed: %s", conn.ID()[:8],
			connOpts.systemUser.Protocol, err)
		errMsg := "unsupported protocol %s"
		errMsg = fmt.Sprintf(errMsg, connOpts.systemUser.Protocol)
		err = fmt.Errorf("%w: %s", ErrUnMatchProtocol, err)
		common.IgnoreErrWriteString(conn, common.WrapperWarn(errMsg))
		return nil, err
	}

	var (
		err          error
		session      *model.Session
		terminalConf model.TerminalConfig
		expireInfo   *model.ExpireInfo
	)

	terminalConf, err = core.GetTerminalConfig()
	if err != nil {
		return nil, fmt.Errorf("get terminal config error: %s", err)
	}

	if !connOpts.asset.IsSupportProtocol(connOpts.systemUser.Protocol) {
		msg := "System user <%s> and asset <%s> protocol are inconsistent."
		msg = fmt.Sprintf(msg, connOpts.systemUser.Username, connOpts.asset.Hostname)
		common.IgnoreErrWriteString(conn, common.WrapperWarn(msg))
		return nil, fmt.Errorf("%w: %s", ErrUnMatchProtocol, msg)
	}

	session = &model.Session{
		ID:           common.UUID(),
		User:         connOpts.user.String(),
		LoginFrom:    conn.LoginFrom(),
		RemoteAddr:   conn.RemoteAddr(),
		Protocol:     connOpts.systemUser.Protocol,
		UserID:       connOpts.user.ID,
		SystemUser:   connOpts.systemUser.Username,
		SystemUserID: connOpts.systemUser.ID,
		Asset:        connOpts.asset.String(),
		AssetID:      connOpts.asset.ID,
	}

	if expireInfo == nil {
		expireInfo, err = core.ValidateAssetConnectPermission(connOpts.user.ID, connOpts.asset.ID)
		if err != nil {
			log.Error.Print(err)
		}
	}

	return &Server{
		ID:       session.ID,
		UserConn: conn,
		core:     core,

		connOpts: connOpts,

		terminalConf: &terminalConf,
		expireInfo:   expireInfo,
		sessionInfo:  session,
		CreateSessionCallback: func() error {
			session.DateStart = time.Now()
			return core.CreateSession(*session)
		},
		ConnectedSuccessCallback: func() error {
			return core.SessionSuccess(session.ID)
		},
		ConnectedFailedCallback: func(err error) error {
			return core.SessionFailed(session.ID, err)
		},
		DisConnectedCallback: func() error {
			return core.SessionDisconnect(session.ID)
		},
	}, nil
}

type Server struct {
	ID       string
	UserConn UserConnection
	core     *core.Core

	connOpts *ConnectionOptions

	terminalConf *model.TerminalConfig
	expireInfo   *model.ExpireInfo

	sessionInfo *model.Session

	cacheSSHConnection *srvconn.SSHConnection

	CreateSessionCallback    func() error
	ConnectedSuccessCallback func() error
	ConnectedFailedCallback  func(err error) error
	DisConnectedCallback     func() error

	keyboardMode int32

	loginTicketId string
}

func (s *Server) Proxy() {
	if err := s.checkRequiredAuth(); err != nil {
		log.Error.Printf("Conn[%s]: check basic auth failed: %s", s.UserConn.ID()[:8], err)
		return
	}
	defer func() {
		if s.cacheSSHConnection != nil {
			_ = s.cacheSSHConnection.Close()
		}
	}()
	if s.CheckPermissionExpired(time.Now()) {
		log.Info.Printf("Conn[%s]: %s has expired to login to %s", s.UserConn.ID()[:8], s.connOpts.user,
			s.connOpts.asset.Name)
		common.IgnoreErrWriteString(s.UserConn, "Your permission to login to this asset has expired")
		common.IgnoreErrWriteString(s.UserConn, common.CharNewLine)
		return
	}
	if !s.checkLoginConfirm() {
		log.Info.Printf("Conn[%s]: check login confirm failed", s.UserConn.ID()[:8])
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	sw := SwitchSession{
		ID:            s.ID,
		MaxIdleTime:   s.terminalConf.MaxIdleTime,
		keepAliveTime: 60,
		ctx:           ctx,
		cancel:        cancel,
		p:             s,
	}
	if err := s.CreateSessionCallback(); err != nil {
		msg := "Connect server failed"
		msg = common.WrapperWarn(msg)
		common.IgnoreErrWriteString(s.UserConn, msg)
		log.Error.Printf("Conn[%s] submit session %s to core server err: %s",
			s.UserConn.ID()[:8], s.ID[:8], msg)
		return
	}
	AddCommonSwitch(&sw)
	defer RemoveCommonSwitch(&sw)
	defer func() {
		if err := s.DisConnectedCallback(); err != nil {
			log.Error.Printf("Conn[%s] update session %s err: %+v", s.UserConn.ID()[:8], s.ID[:8], err)
		}
	}()
	var proxyAddr *net.TCPAddr
	srvCon, err := s.getServerConn(proxyAddr)
	if err != nil {
		log.Error.Print(err)
		s.sendConnectErrorMsg(err)
		if err2 := s.ConnectedFailedCallback(err); err2 != nil {
			log.Error.Printf("Conn[%s] update session err: %s", s.UserConn.ID()[:8], err2)
		}
		return
	}
	defer srvCon.Close()

	log.Info.Printf("Conn[%s] create session %s success", s.UserConn.ID()[:8], s.ID[:8])
	if err2 := s.ConnectedSuccessCallback(); err2 != nil {
		log.Error.Printf("Conn[%s] update session %s err: %s", s.UserConn.ID()[:8], s.ID[:8], err2)
	}
	common.IgnoreErrWriteWindowTitle(s.UserConn, s.connOpts.TerminalTitle())
	if err = sw.Bridge(s.UserConn, srvCon); err != nil {
		log.Error.Print(err)
	}
}

func (s *Server) IsKeyboardMode() bool {
	return atomic.LoadInt32(&s.keyboardMode) == 1
}

func (s *Server) setKeyBoardMode() {
	atomic.StoreInt32(&s.keyboardMode, 1)
}

func (s *Server) resetKeyboardMode() {
	atomic.StoreInt32(&s.keyboardMode, 0)
}

func (s *Server) sendConnectErrorMsg(err error) {
	msg := fmt.Sprintf("%s error: %s", s.connOpts.ConnectMsg(),
		s.ConvertErrorToReadableMsg(err))
	common.IgnoreErrWriteString(s.UserConn, msg)
	common.IgnoreErrWriteString(s.UserConn, common.CharNewLine)
	log.Error.Print(msg)
	password := s.connOpts.systemUser.Password
	if password != "" {
		msg2 := fmt.Sprintf("Try password: %s", password[:2]+strings.Repeat("*", 8))
		log.Info.Print(msg2)
	}
}

func (s *Server) CheckPermissionExpired(now time.Time) bool {
	return s.expireInfo.IsExpired(now)
}

func (s *Server) checkRequiredAuth() error {
	switch s.connOpts.systemUser.Protocol {
	case srvconn.ProtocolSSH:
		if err := s.getUsernameIfNeed(); err != nil {
			msg := common.WrapperWarn("Get auth username failed")
			common.IgnoreErrWriteString(s.UserConn, msg)
			return err
		}
		if s.checkReuseSSHClient() {
			if cacheConn, ok := s.getCacheSSHConn(); ok {
				s.cacheSSHConnection = cacheConn
				return nil
			}
			log.Debug.Printf("Conn[%s] did not found cache ssh client(%s@%s)",
				s.UserConn.ID()[:8], s.connOpts.systemUser.Username, s.connOpts.asset.Hostname)
		}

		if s.connOpts.systemUser.PrivateKey == "" {
			if err := s.getAuthPasswordIfNeed(); err != nil {
				msg := common.WrapperWarn("Get auth password failed")
				common.IgnoreErrWriteString(s.UserConn, msg)
				return err
			}
		}
	default:
		return ErrNoAuthInfo
	}
	return nil
}

func (s *Server) getUsernameIfNeed() (err error) {
	if s.connOpts.systemUser.Username == "" {
		log.Info.Printf("Conn[%s] need manuel input system user username", s.UserConn.ID()[:8])
		var username string
		term := common.NewTerminal(s.UserConn, "username: ")
		for {
			username, err = term.ReadLine()
			if err != nil {
				return err
			}
			username = strings.TrimSpace(username)
			if username != "" {
				break
			}
		}
		s.connOpts.systemUser.Username = username
		log.Info.Printf("Conn[%s] get username from user input: %s", s.UserConn.ID()[:8], username)
	}
	return
}

func (s *Server) getAuthPasswordIfNeed() (err error) {
	var line string
	if s.connOpts.systemUser.Password == "" {
		term := common.NewTerminal(s.UserConn, "password: ")
		if s.connOpts.systemUser.Username != "" {
			line, err = term.ReadPassword(fmt.Sprintf("%s's password: ", s.connOpts.systemUser.Username))
		} else {
			line, err = term.ReadPassword("password: ")
		}

		if err != nil {
			log.Error.Printf("Conn[%s] get password from user err: %s", s.UserConn.ID()[:8], err.Error())
			return err
		}
		s.connOpts.systemUser.Password = line
		log.Info.Printf("Conn[%s] get password from user input", s.UserConn.ID()[:8])
	}
	return nil
}

func (s *Server) getServerConn(proxyAddr *net.TCPAddr) (srvconn.ServerConnection, error) {
	if s.cacheSSHConnection != nil {
		return s.cacheSSHConnection, nil
	}
	done := make(chan struct{})
	defer func() {
		common.IgnoreErrWriteString(s.UserConn, "\r\n")
		close(done)
	}()
	go s.sendConnectingMsg(done)
	switch s.connOpts.systemUser.Protocol {
	case srvconn.ProtocolSSH:
		return s.getSSHConn()
	default:
		return nil, ErrUnMatchProtocol
	}
}

func (s *Server) getSSHConn() (srvConn *srvconn.SSHConnection, err error) {
	loginSystemUser := s.connOpts.systemUser
	key := srvconn.MakeReuseSSHClientKey(s.connOpts.user.ID, s.connOpts.asset.ID, loginSystemUser.ID,
		s.connOpts.asset.IP, loginSystemUser.Username)
	timeout := config.GlobalConfig.SSHTimeout
	sshAuthOpts := make([]srvconn.SSHClientOption, 0, 6)
	sshAuthOpts = append(sshAuthOpts, srvconn.SSHClientUsername(loginSystemUser.Username))
	sshAuthOpts = append(sshAuthOpts, srvconn.SSHClientHost(s.connOpts.asset.IP))
	sshAuthOpts = append(sshAuthOpts, srvconn.SSHClientPort(s.connOpts.asset.ProtocolPort(loginSystemUser.Protocol)))
	sshAuthOpts = append(sshAuthOpts, srvconn.SSHClientPassword(loginSystemUser.Password))
	sshAuthOpts = append(sshAuthOpts, srvconn.SSHClientTimeout(timeout))
	if loginSystemUser.PrivateKey != "" {
		// 先使用 password 解析 PrivateKey
		if signer, err1 := gossh.ParsePrivateKeyWithPassphrase([]byte(loginSystemUser.PrivateKey),
			[]byte(loginSystemUser.Password)); err1 == nil {
			sshAuthOpts = append(sshAuthOpts, srvconn.SSHClientPrivateAuth(signer))
		} else {
			// 如果之前使用password解析失败，则去掉 password, 尝试直接解析 PrivateKey 防止错误的passphrase
			if signer, err1 = gossh.ParsePrivateKey([]byte(loginSystemUser.PrivateKey)); err1 == nil {
				sshAuthOpts = append(sshAuthOpts, srvconn.SSHClientPrivateAuth(signer))
			}
		}
	}
	password := loginSystemUser.Password
	privateKey := loginSystemUser.PrivateKey
	kb := srvconn.SSHClientKeyboardAuth(func(user, instruction string,
		questions []string, echos []bool) (answers []string, err error) {
		s.setKeyBoardMode()
		termReader := common.NewTerminal(s.UserConn, "")
		common.IgnoreErrWriteString(s.UserConn, "\r\n")
		ans := make([]string, len(questions))
		for i := range questions {
			q := questions[i]
			termReader.SetPrompt(questions[i])
			log.Debug.Printf("Conn[%s] keyboard auth question [ %s ]", s.UserConn.ID()[:8], q)
			if strings.Contains(strings.ToLower(q), "password") {
				if privateKey != "" || password != "" {
					ans[i] = password
					continue
				}
			}
			line, err2 := termReader.ReadLine()
			if err2 != nil {
				log.Error.Printf("Conn[%s] keyboard auth read err: %s", s.UserConn.ID()[:8], err2)
			}
			ans[i] = line
		}
		s.resetKeyboardMode()
		return ans, nil
	})
	sshAuthOpts = append(sshAuthOpts, kb)
	sshClient, err := srvconn.NewSSHClient(sshAuthOpts...)
	if err != nil {
		log.Error.Printf("Get new ssh client err: %s", err)
		return nil, err
	}
	srvconn.AddClientCache(key, sshClient)
	sess, err := sshClient.AcquireSession()
	if err != nil {
		log.Error.Printf("SSH client(%s) start session err %s", sshClient, err)
		return nil, err
	}
	pty := s.UserConn.Pty()
	sshConnectOpts := make([]srvconn.SSHOption, 0, 6)
	sshConnectOpts = append(sshConnectOpts, srvconn.SSHTerm(pty.Term))
	sshConnectOpts = append(sshConnectOpts, srvconn.SSHPtyWin(srvconn.Windows{
		Width:  pty.Window.Width,
		Height: pty.Window.Height,
	}))

	sshConn, err := srvconn.NewSSHConnection(sess, sshConnectOpts...)
	if err != nil {
		_ = sess.Close()
		sshClient.ReleaseSession(sess)
		return nil, err
	}

	go func() {
		_ = sess.Wait()
		sshClient.ReleaseSession(sess)
		log.Info.Printf("SSH client(%s) shell connection release", sshClient)
	}()
	return sshConn, nil

}

func (s *Server) sendConnectingMsg(done chan struct{}) {
	delay := 0.0
	maxDelay := 5 * 60.0 // 最多执行五分钟
	msg := fmt.Sprintf("%s  %.1f", s.connOpts.ConnectMsg(), delay)
	common.IgnoreErrWriteString(s.UserConn, msg)
	var activeFlag bool
	for delay < maxDelay {
		select {
		case <-done:
			return
		default:
			if s.IsKeyboardMode() {
				activeFlag = true
				break
			}
			if activeFlag {
				common.IgnoreErrWriteString(s.UserConn, common.CharClear)
				msg = fmt.Sprintf("%s  %.1f", s.connOpts.ConnectMsg(), delay)
				common.IgnoreErrWriteString(s.UserConn, msg)
				activeFlag = false
				break
			}
			delayS := fmt.Sprintf("%.1f", delay)
			data := strings.Repeat("\x08", len(delayS)) + delayS
			common.IgnoreErrWriteString(s.UserConn, data)
		}
		time.Sleep(100 * time.Millisecond)
		delay += 0.1
	}
}

func (s *Server) checkReuseSSHClient() bool {
	if config.GetConf().ReuseConnection {
		platformMatched := s.connOpts.asset.Platform == "Linux"
		protocolMatched := s.connOpts.systemUser.Protocol == model.ProtocolSSH
		return platformMatched && protocolMatched
	}
	return false
}

func (s *Server) getCacheSSHConn() (srvConn *srvconn.SSHConnection, ok bool) {
	keyId := srvconn.MakeReuseSSHClientKey(s.connOpts.user.ID, s.connOpts.asset.ID,
		s.connOpts.systemUser.ID, s.connOpts.asset.IP, s.connOpts.systemUser.Username)
	sshClient, ok := srvconn.GetClientFromCache(keyId)
	if !ok {
		return nil, ok
	}
	sess, err := sshClient.AcquireSession()
	if err != nil {
		log.Error.Printf("Cache ssh client new session failed: %s", err)
		return nil, false
	}
	pty := s.UserConn.Pty()
	cacheConn, err := srvconn.NewSSHConnection(sess,
		srvconn.SSHPtyWin(srvconn.Windows{
			Width:  pty.Window.Width,
			Height: pty.Window.Height,
		}), srvconn.SSHTerm(pty.Term))
	if err != nil {
		log.Error.Printf("Cache ssh session failed: %s", err)
		_ = sess.Close()
		sshClient.ReleaseSession(sess)
		return nil, false
	}
	reuseMsg := fmt.Sprintf("Reuse SSH connections (%s@%s) [Number of connections: %d]",
		s.connOpts.systemUser.Username, s.connOpts.asset.IP, sshClient.RefCount())
	common.IgnoreErrWriteString(s.UserConn, reuseMsg+"\r\n")
	go func() {
		_ = sess.Wait()
		sshClient.ReleaseSession(sess)
		log.Info.Printf("Reuse SSH client(%s) shell connection release", sshClient)
	}()
	return cacheConn, true
}

func (s *Server) checkLoginConfirm() bool {
	opts := make([]auth.ConfirmOption, 0, 4)
	opts = append(opts, auth.ConfirmWithUser(s.connOpts.user))
	opts = append(opts, auth.ConfirmWithSystemUser(s.connOpts.systemUser))

	var targetId, targetName string
	switch s.connOpts.systemUser.Protocol {
	case srvconn.ProtocolSSH:
		targetId = s.connOpts.asset.ID
		targetName = s.connOpts.asset.Name
	}
	opts = append(opts, auth.ConfirmWithAssetID(targetId), auth.ConfirmWithAssetName(targetName))
	confirmSrv := auth.NewLoginConfirm(s.core, opts...)
	ok := s.validateLoginConfirm(&confirmSrv, s.UserConn)
	s.loginTicketId = confirmSrv.GetTicketId()
	return ok
}

func (s *Server) ConvertErrorToReadableMsg(e error) string {
	if e == nil {
		return ""
	}
	errMsg := e.Error()
	if strings.Contains(errMsg, "unable to authenticate") || strings.Contains(errMsg, "failed login") {
		return "Authentication failed"
	}
	if strings.Contains(errMsg, "connection refused") {
		return "Connection refused"
	}
	if strings.Contains(errMsg, "i/o timeout") {
		return "i/o timeout"
	}
	if strings.Contains(errMsg, "No route to host") {
		return "No route to host"
	}
	if strings.Contains(errMsg, "network is unreachable") {
		return "network is unreachable"
	}
	return errMsg
}

func (s *Server) GetReplayRecorder() *ReplyRecorder {
	pty := s.UserConn.Pty()
	info := &ReplyInfo{
		Width:     pty.Window.Width,
		Height:    pty.Window.Height,
		TimeStamp: time.Now(),
	}
	recorder, err := NewReplayRecord(s.ID, s.connOpts.user.Username,
		s.connOpts.asset.Name, info)
	if err != nil {
		log.Error.Print(err)
	}
	return recorder
}
