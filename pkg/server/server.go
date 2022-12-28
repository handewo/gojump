package server

import (
	"context"
	"net"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gliderlabs/ssh"
	"github.com/handewo/gojump/pkg/config"
	"github.com/handewo/gojump/pkg/core"
	"github.com/handewo/gojump/pkg/log"
	"github.com/handewo/gojump/pkg/model"
	"github.com/pires/go-proxyproto"
)

type server struct {
	terminalConf atomic.Value
	core         *core.Core
	srv          *ssh.Server
	sync.Mutex
}

func (s *server) updateTermCfgPeriodcally() {
	for {
		time.Sleep(time.Minute)
		conf, err := s.core.GetTerminalConfig()
		if err != nil {
			log.Error.Printf("Update terminal config failed: %s", err)
			continue
		}
		s.UpdateTerminalConfig(conf)
	}
}

func (s *server) UpdateTerminalConfig(conf model.TerminalConfig) {
	s.terminalConf.Store(conf)
}

func (s *server) GetTerminalConfig() model.TerminalConfig {
	return s.terminalConf.Load().(model.TerminalConfig)
}

func Run(cfg string) {
	config.Initial(cfg)

	log.SetLogFile(config.GlobalConfig.LogFile)
	log.SetLogLevel(config.GlobalConfig.LogLevel)

	gracefulStop := make(chan os.Signal, 1)
	signal.Notify(gracefulStop, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	core := core.NewCore()
	srv := NewServer(core)
	srv.initSSHServer()
	go srv.Serve()
	defer srv.Shutdown()
	<-gracefulStop
}

func NewServer(c *core.Core) *server {
	terminalConf, err := c.GetTerminalConfig()
	if err != nil {
		log.Fatal.Fatal(err)
	}
	srv := server{
		core: c,
	}
	srv.UpdateTerminalConfig(terminalConf)
	go srv.updateTermCfgPeriodcally()
	return &srv
}

func (s *server) Serve() {
	log.Info.Printf("Start SSH server at %s", s.srv.Addr)
	ln, err := net.Listen("tcp", s.srv.Addr)
	if err != nil {
		log.Fatal.Print(err)
	}
	proxyListener := &proxyproto.Listener{Listener: ln}
	if err := s.srv.Serve(proxyListener); err != nil {
		log.Fatal.Print(err)
	}
}

func (s *server) Shutdown() {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer log.Close()
	defer s.core.Close()
	defer cancelFunc()
	if err := s.srv.Shutdown(ctx); err != nil {
		log.Fatal.Print(err)
	}
}

func (s *server) initSSHServer() {
	srv := &ssh.Server{
		Addr: s.GetSSHAddr(),
		PasswordHandler: func(ctx ssh.Context, password string) bool {
			return s.PasswordAuth(ctx, password)
		},
		PublicKeyHandler: func(ctx ssh.Context, key ssh.PublicKey) bool {
			return s.PublicKeyAuth(ctx, key)
		},
		HostSigners: []ssh.Signer{s.GetSSHSigner()},
		Handler:     s.SessionHandler,
		ChannelHandlers: map[string]ssh.ChannelHandler{
			"session": ssh.DefaultSessionHandler,
		},
	}
	s.srv = srv
}
