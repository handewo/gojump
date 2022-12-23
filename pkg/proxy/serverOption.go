package proxy

import (
	"fmt"

	"github.com/handewo/gojump/pkg/model"
	"github.com/handewo/gojump/pkg/srvconn"
)

type ConnectionOption func(options *ConnectionOptions)

func ConnectUser(user *model.User) ConnectionOption {
	return func(opts *ConnectionOptions) {
		opts.user = user
	}
}

func ConnectSystemUser(systemUser *model.SystemUser) ConnectionOption {
	return func(opts *ConnectionOptions) {
		opts.systemUser = systemUser
	}
}

func ConnectAsset(asset *model.Asset) ConnectionOption {
	return func(opts *ConnectionOptions) {
		opts.asset = asset
	}
}

type ConnectionOptions struct {
	user       *model.User
	systemUser *model.SystemUser

	asset *model.Asset
}

func (opts *ConnectionOptions) TerminalTitle() string {
	title := ""
	switch opts.systemUser.Protocol {
	case srvconn.ProtocolSSH:
		title = fmt.Sprintf("%s://%s@%s",
			opts.systemUser.Protocol,
			opts.systemUser.Username,
			opts.asset.IP)
	}
	return title
}

func (opts *ConnectionOptions) ConnectMsg() string {
	msg := ""
	switch opts.systemUser.Protocol {
	case srvconn.ProtocolSSH:
		msg = fmt.Sprintf("Connecting to %s@%s", opts.systemUser.Username, opts.asset.IP)
	}
	return msg
}
