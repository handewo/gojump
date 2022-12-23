package srvconn

import (
	"errors"
	"io"

	gossh "golang.org/x/crypto/ssh"
)

func NewSSHConnection(sess *gossh.Session, opts ...SSHOption) (*SSHConnection, error) {
	if sess == nil {
		return nil, errors.New("ssh session is nil")
	}
	options := &SSHOptions{
		win: Windows{
			Width:  80,
			Height: 120,
		},
		term: "xterm",
	}
	for _, setter := range opts {
		setter(options)
	}
	modes := gossh.TerminalModes{
		gossh.ECHO:          1,     // enable echoing
		gossh.TTY_OP_ISPEED: 14400, // input speed = 14.4 kbaud
		gossh.TTY_OP_OSPEED: 14400, // output speed = 14.4 kbaud
	}
	err := sess.RequestPty(options.term, options.win.Height, options.win.Width, modes)
	if err != nil {
		return nil, err
	}
	stdin, err := sess.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := sess.StdoutPipe()
	if err != nil {
		return nil, err
	}
	conn := &SSHConnection{
		session: sess,
		stdin:   stdin,
		stdout:  stdout,
		options: options,
	}
	err = sess.Shell()
	if err != nil {
		_ = sess.Close()
		return nil, err
	}
	return conn, nil
}

type SSHConnection struct {
	session *gossh.Session
	stdin   io.Writer
	stdout  io.Reader
	options *SSHOptions
}

func (sc *SSHConnection) SetWinSize(w, h int) error {
	return sc.session.WindowChange(h, w)
}

func (sc *SSHConnection) Read(p []byte) (n int, err error) {
	return sc.stdout.Read(p)
}

func (sc *SSHConnection) Write(p []byte) (n int, err error) {
	return sc.stdin.Write(p)
}

func (sc *SSHConnection) Close() (err error) {
	return sc.session.Close()
}

func (sc *SSHConnection) KeepAlive() error {
	_, err := sc.session.SendRequest("keepalive@openssh.com", false, nil)
	return err
}

type SSHOption func(*SSHOptions)

type SSHOptions struct {
	win  Windows
	term string
}

func SSHPtyWin(win Windows) SSHOption {
	return func(opt *SSHOptions) {
		opt.win = win
	}
}

func SSHTerm(termType string) SSHOption {
	return func(opt *SSHOptions) {
		opt.term = termType
	}
}
