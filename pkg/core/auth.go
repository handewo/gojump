package core

import (
	"errors"
	"fmt"
	"time"

	"github.com/handewo/gojump/pkg/common"
	"github.com/handewo/gojump/pkg/config"
	"github.com/handewo/gojump/pkg/log"
	"github.com/handewo/gojump/pkg/model"
	uuid "github.com/satori/go.uuid"
)

func (c *Core) CheckIfNeedAssetLoginConfirm(userId, username, assetId, assetName,
	sysUsername string) (res model.AssetLoginTicketInfo, err error) {
	need := true
	err = c.db.QueryOneField(&need,
		"SELECT needconfirm FROM ASSETUSERINFO WHERE userid = ? and assetid = ?", userId, assetId)
	if err != nil {
		log.Error.Printf("get needconfirm falied, %s", err)
		return model.AssetLoginTicketInfo{}, err
	}
	if !need {
		return model.AssetLoginTicketInfo{NeedConfirm: need}, nil
	}
	ticketId := uuid.NewV4().String()
	date := time.Now()
	cl := model.LoginTicket{
		TicketId:        ticketId,
		State:           model.TicketOpen,
		Approver:        "",
		ApplicationDate: date.Format(common.LogFormat),
		ApproveDate:     "",
		Username:        username,
		AssetName:       assetName,
		SysUsername:     sysUsername,
	}
	err = c.db.InsertData("INSERT INTO LOGINTICKET VALUES ?", &cl)
	if err != nil {
		log.Error.Printf("insert into LOGINTICKET falied, %s", err)
		return model.AssetLoginTicketInfo{}, err
	}

	reviewers, err := c.getReviewers()
	if err != nil {
		log.Error.Printf("get reviewers falied, %s", err)
		return model.AssetLoginTicketInfo{}, err
	}
	return model.AssetLoginTicketInfo{
		TicketId:    ticketId,
		NeedConfirm: need,
		Reviewers:   reviewers}, nil
}

func (c *Core) CheckConfirmStatusByRequestInfo(ticketId string) (model.TicketState, error) {
	var approver, state string

	err := c.db.QueryTwoField(&approver, &state, "SELECT approver, state FROM LOGINTICKET where ticketid = ?", ticketId)
	if err != nil {
		return model.TicketState{State: "error"}, err
	}
	return model.TicketState{
		Approver: approver,
		State:    state}, nil
}

func (c *Core) CancelConfirmByRequestInfo(ticketId string) error {
	err := c.UpdateTicketState(ticketId, model.TicketClosed, "")
	return err
}

func (c *Core) UpdateTicketState(ticketId string, state string, user string) error {
	date := time.Now().Format(common.LogFormat)
	err := c.db.UpdateData("UPDATE LOGINTICKET SET state = ?, approver = ?, approvedate = ? WHERE ticketid = ? AND state = ?",
		state, user, date, ticketId, model.TicketOpen)
	return err
}

func (c *Core) getReviewers() ([]string, error) {
	res, err := c.db.QueryOneFieldMutilRows("SELECT username from USER WHERE role = ?", "admin")
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Core) QueryPengdingLoginTicket() ([]string, error) {
	v, err := c.db.QueryStructs(model.LoginTicketType, "SELECT * FROM LOGINTICKET WHERE state = ?",
		model.TicketOpen)
	if err != nil {
		return nil, err
	}

	lgtik, ok := v.([]model.LoginTicket)
	if !ok {
		return nil, errors.New("invalid value type")
	}

	ticks := make([]string, 0, 5)
	for _, v := range lgtik {
		ticks = append(ticks, fmt.Sprintf("%s|%s|%10s|%10s|%10s", v.TicketId, v.ApplicationDate, v.Username,
			v.AssetName, v.SysUsername))
	}
	return ticks, err
}

func (c *Core) QueryLoginTicket() ([]string, error) {
	v, err := c.db.QueryStructs(model.LoginTicketType, "SELECT * FROM LOGINTICKET")
	if err != nil {
		return nil, err
	}

	lgtik, ok := v.([]model.LoginTicket)
	if !ok {
		return nil, errors.New("invalid value type")
	}

	ticks := make([]string, 0, 10)
	for _, v := range lgtik {
		ticks = append(ticks, fmt.Sprintf("%s|%s|%10s|%10s|%10s|%10s|%19s|%10s", v.TicketId,
			v.ApplicationDate, v.Username, v.AssetName, v.SysUsername, v.Approver, v.ApproveDate, v.State))
	}
	return ticks, nil
}

func (c *Core) AuthenticationLog(username, authMethod, remoteAddr string) {
	date := time.Now().Format(common.LogFormat)
	lg := model.UserLog{
		Datetime: date,
		Type:     "auth",
		User:     username,
		Log:      fmt.Sprintf("authenticate successfully from %s using %s", remoteAddr, authMethod),
	}
	err := c.db.InsertData("INSERT INTO USERLOG VALUES ?", &lg)
	if err != nil {
		log.Error.Printf("insert authentication log failed, %s", err)
	}
}

func (c *Core) LimitTryLogin(user string) {
	c.loginLock.Lock()
	count, ok := c.tryLoginCount[user]
	c.tryLoginCount[user] = count + 1
	c.loginLock.Unlock()
	log.Debug.Printf("%s has tried login %d times", user, count+1)
	if !ok {
		go func() {
			blockTime := time.Duration(config.GlobalConfig.LoginBlockTime) * time.Minute
			timer := time.NewTimer(blockTime)
			<-timer.C
			c.loginLock.Lock()
			delete(c.tryLoginCount, user)
			c.loginLock.Unlock()
			log.Info.Printf("%s login limit is reset", user)
		}()
	}
}

func (c *Core) UserIsBlocked(user string) bool {
	c.loginLock.RLock()
	count := c.tryLoginCount[user]
	c.loginLock.RUnlock()
	return count >= config.GlobalConfig.MaxTryLogin
}
