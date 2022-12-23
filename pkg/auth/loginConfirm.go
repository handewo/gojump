package auth

import (
	"context"
	"time"

	"github.com/handewo/gojump/pkg/core"
	"github.com/handewo/gojump/pkg/log"
	"github.com/handewo/gojump/pkg/model"
)

type connectionConfirmOption struct {
	user       *model.User
	systemUser *model.SystemUser

	assetID   string
	assetName string
}

func NewLoginConfirm(core *core.Core, opts ...ConfirmOption) LoginConfirmService {
	var option connectionConfirmOption
	for _, setter := range opts {
		setter(&option)
	}
	return LoginConfirmService{option: &option, core: core}
}

type LoginConfirmService struct {
	core *core.Core

	option *connectionConfirmOption

	reviewers []string

	approver string
	ticketId string
}

func (c *LoginConfirmService) CheckIsNeedLoginConfirm() (bool, error) {
	userid := c.option.user.ID
	username := c.option.user.Username
	systemUsername := c.option.systemUser.Username
	assetId := c.option.assetID
	assetName := c.option.assetName
	res, err := c.core.CheckIfNeedAssetLoginConfirm(userid, username, assetId, assetName,
		systemUsername)
	if err != nil {
		return false, err
	}
	c.ticketId = res.TicketId
	c.reviewers = res.Reviewers
	return res.NeedConfirm, nil
}

func (c *LoginConfirmService) WaitLoginConfirm(ctx context.Context) Status {
	return c.waitConfirmFinish(ctx)
}

func (c *LoginConfirmService) GetReviewers() []string {
	reviewers := make([]string, len(c.reviewers))
	copy(reviewers, c.reviewers)
	return reviewers
}

func (c *LoginConfirmService) GetApprover() string {
	return c.approver
}

func (c *LoginConfirmService) GetTicketId() string {
	return c.ticketId
}

func (c *LoginConfirmService) waitConfirmFinish(ctx context.Context) Status {
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			c.cancelConfirm()
			return StatusCancel
		case <-t.C:
			statusRes, err := c.core.CheckConfirmStatusByRequestInfo(c.ticketId)
			if err != nil {
				log.Error.Printf("Check confirm status err: %s", err.Error())
				continue
			}
			switch statusRes.State {
			case model.TicketOpen:
				continue
			case model.TicketApproved:
				c.approver = statusRes.Approver
				return StatusApprove
			case model.TicketRejected, model.TicketClosed:
				c.approver = statusRes.Approver
				return StatusReject
			default:
				log.Error.Printf("Receive unknown login confirm status %s",
					statusRes.State)
			}
		}
	}
}

func (c *LoginConfirmService) cancelConfirm() {
	if err := c.core.CancelConfirmByRequestInfo(c.ticketId); err != nil {
		log.Error.Printf("Cancel confirm request err: %s", err.Error())
	}
}

type Status int

const (
	StatusApprove Status = iota + 1
	StatusReject
	StatusCancel
)

type ConfirmOption func(*connectionConfirmOption)

func ConfirmWithUser(user *model.User) ConfirmOption {
	return func(option *connectionConfirmOption) {
		option.user = user
	}
}

func ConfirmWithSystemUser(sysUser *model.SystemUser) ConfirmOption {
	return func(option *connectionConfirmOption) {
		option.systemUser = sysUser
	}
}

func ConfirmWithAssetID(AssetId string) ConfirmOption {
	return func(option *connectionConfirmOption) {
		option.assetID = AssetId
	}
}

func ConfirmWithAssetName(AssetName string) ConfirmOption {
	return func(option *connectionConfirmOption) {
		option.assetName = AssetName
	}
}
