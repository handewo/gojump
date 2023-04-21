package handler

import (
	"fmt"

	"github.com/gliderlabs/ssh"
	"github.com/handewo/gojump/pkg/common"
	"github.com/handewo/gojump/pkg/core"
	"github.com/handewo/gojump/pkg/log"
	"github.com/handewo/gojump/pkg/model"
	"github.com/handewo/gojump/pkg/proxy"
)

type DirectHandler struct {
	term        *common.Terminal
	sess        ssh.Session
	wrapperSess *WrapperSession
	core        *core.Core
	user        *model.User
	asset       *model.Asset
	sysUser     *model.SystemUser
	directLogin map[string]string
}

func NewDirectHandler(sess ssh.Session, core *core.Core, user *model.User,
	directLogin map[string]string) (*DirectHandler, error) {
	var (
		wrapperSess *WrapperSession
		term        *common.Terminal
	)
	asset, err := core.GetAssetByName(directLogin["asset"])
	if err != nil {
		return nil, err
	}
	if asset.ID == "" {
		return nil, fmt.Errorf("asset %s not found", directLogin["asset"])
	}

	sysUsers, err := core.GetSystemUsersByUserIdAndAssetId(user.ID, asset.ID)
	if err != nil {
		return nil, err
	}
	var sysUser model.SystemUser
	for _, v := range sysUsers {
		if v.Username == directLogin["sysuser"] {
			sysUser = v
		}
	}
	if sysUser.ID == "" {
		return nil, fmt.Errorf("system user %s not found", directLogin["sysuser"])
	}

	wrapperSess = NewWrapperSession(sess)
	term = common.NewTerminal(wrapperSess, "Opt> ")
	d := &DirectHandler{
		sess:        sess,
		core:        core,
		directLogin: directLogin,
		wrapperSess: wrapperSess,
		term:        term,
		user:        user,
		asset:       &asset,
		sysUser:     &sysUser,
	}
	return d, nil
}

func (d *DirectHandler) Proxy() error {
	_, winChan, _ := d.sess.Pty()
	go d.WatchWinSizeChange(winChan)

	srv, err := proxy.NewServer(d.wrapperSess,
		d.core,
		proxy.ConnectUser(d.user),
		proxy.ConnectAsset(d.asset),
		proxy.ConnectSystemUser(d.sysUser),
	)
	if err != nil {
		return err
	}
	srv.Proxy()
	log.Info.Printf("Request %s: asset %s proxy end", d.wrapperSess.Uuid, d.asset.Name)
	return nil
}

func (d *DirectHandler) WatchWinSizeChange(winChan <-chan ssh.Window) {
	defer log.Debug.Printf("Request %s: Windows change watch close", d.wrapperSess.Uuid)
	for {
		select {
		case <-d.sess.Context().Done():
			return
		case win, ok := <-winChan:
			if !ok {
				return
			}
			d.wrapperSess.SetWin(win)
			log.Debug.Printf("Term window size change: %d*%d", win.Height, win.Width)
			_ = d.term.SetSize(win.Width, win.Height)
		}
	}
}
