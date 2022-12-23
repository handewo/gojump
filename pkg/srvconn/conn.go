package srvconn

import (
	"errors"
	"fmt"
	"io"
)

const (
	ProtocolSSH = "ssh"
)

var (
	ErrUnSupportedProtocol = errors.New("unsupported protocol")
)

func IsSupportedProtocol(p string) error {
	if p != ProtocolSSH {
		return ErrUnSupportedProtocol
	}
	return nil
}

type ServerConnection interface {
	io.ReadWriteCloser
	SetWinSize(width, height int) error
	KeepAlive() error
}

type Windows struct {
	Width  int
	Height int
}

func MakeReuseSSHClientKey(userId, assetId, systemUserId,
	ip, username string) string {
	return fmt.Sprintf("%s_%s_%s_%s_%s", userId, assetId,
		systemUserId, ip, username)
}

var sshManager = newSSHManager()

func GetClientFromCache(key string) (client *SSHClient, ok bool) {
	return sshManager.getClientFromCache(key)
}
