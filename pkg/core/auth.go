package core

import (
	"fmt"
	"time"

	"github.com/genjidb/genji"
	"github.com/genjidb/genji/document"
	"github.com/genjidb/genji/types"
	"github.com/handewo/gojump/pkg/common"
	"github.com/handewo/gojump/pkg/log"
	"github.com/handewo/gojump/pkg/model"
	uuid "github.com/satori/go.uuid"
)

func (c *Core) CheckIfNeedAssetLoginConfirm(userId, username, assetId, assetName,
	sysUsername string) (res model.AssetLoginTicketInfo, err error) {
	need := true
	err = queryOneFieldFromDb(c.db, &need,
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
	err = InsertData(c.db, "INSERT INTO LOGINTICKET VALUES ?", &cl)
	if err != nil {
		log.Error.Printf("insert into LOGINTICKET falied, %s", err)
		return model.AssetLoginTicketInfo{}, err
	}

	reviewers, err := getReviewers(c.db)
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

	res, err := c.db.Query("SELECT approver, state FROM LOGINTICKET where ticketid = ?", ticketId)
	if err != nil {
		return model.TicketState{State: "error"}, err
	}
	defer res.Close()
	err = res.Iterate(func(d types.Document) error {
		err = document.Scan(d, &approver, &state)
		return err
	})
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
	err := c.db.Exec("UPDATE LOGINTICKET SET state = ?, approver = ?, approvedate = ? WHERE ticketid = ? AND state = ?",
		state, user, date, ticketId, model.TicketOpen)
	return err
}

func getReviewers(db *genji.DB) ([]string, error) {
	res, err := db.Query("SELECT username from USER WHERE role = ?", "admin")
	if err != nil {
		return nil, err
	}
	defer res.Close()
	reviewers := make([]string, 0, 5)
	err = res.Iterate(func(d types.Document) error {
		var r string
		err = document.Scan(d, &r)
		if err != nil {
			return err
		}
		reviewers = append(reviewers, r)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return reviewers, nil
}

func (c *Core) QueryPengdingLoginTicket() ([]string, error) {
	ticks := make([]string, 0, 5)
	res, err := c.db.Query(`SELECT ticketid,applicationdate,username,assetname,
	sysusername FROM LOGINTICKET WHERE state = ?`,
		model.TicketOpen)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	err = res.Iterate(func(d types.Document) error {
		var id, dt, un, an, sn string
		err = document.Scan(d, &id, &dt, &un, &an, &sn)
		if err != nil {
			return err
		}
		ticks = append(ticks, fmt.Sprintf("%s|%s|%10s|%10s|%10s", id, dt, un, an, sn))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ticks, err
}

func (c *Core) QueryLoginTicket() ([]string, error) {
	ticks := make([]string, 0, 5)
	res, err := c.db.Query(`SELECT ticketid,applicationdate,username,assetname,sysusername,
	approver,approvedate,state FROM LOGINTICKET`)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	err = res.Iterate(func(d types.Document) error {
		var id, dt, un, an, sn, ar, ad, st string
		err = document.Scan(d, &id, &dt, &un, &an, &sn, &ar, &ad, &st)
		if err != nil {
			return err
		}
		ticks = append(ticks, fmt.Sprintf("%s|%s|%10s|%10s|%10s|%10s|%19s|%10s",
			id, dt, un, an, sn, ar, ad, st))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ticks, err
}

func (c *Core) AuthenticationLog(username, authMethod, remoteAddr string) {
	date := time.Now().Format(common.LogFormat)
	lg := model.UserLog{
		Datetime: date,
		Type:     "auth",
		User:     username,
		Log:      fmt.Sprintf("authenticate successfully from %s using %s", remoteAddr, authMethod),
	}
	err := InsertData(c.db, model.InsertUserLog, &lg)
	if err != nil {
		log.Error.Printf("insert authentication log failed, %s", err)
	}
}
