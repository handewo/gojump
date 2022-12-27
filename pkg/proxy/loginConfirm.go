package proxy

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/handewo/gojump/pkg/auth"
	"github.com/handewo/gojump/pkg/common"
	"github.com/handewo/gojump/pkg/log"
)

func (s *Server) validateLoginConfirm(srv *auth.LoginConfirmService, userConn UserConnection) bool {
	ok, err := srv.CheckIsNeedLoginConfirm()
	if err != nil {
		log.Error.Printf("Conn[%s] validate login confirm err: %s",
			userConn.ID()[:8], err.Error())
		msg := "validate Login confirm err: Core failed"
		common.IgnoreErrWriteString(userConn, msg)
		common.IgnoreErrWriteString(userConn, common.CharNewLine)
		return false
	}
	if !ok {
		log.Debug.Printf("Conn[%s] no need login confirm", userConn.ID()[:8])
		return true
	}

	ctx, cancelFunc := context.WithCancel(userConn.Context())
	term := common.NewTerminal(userConn, "")
	defer userConn.Close()
	go func() {
		defer cancelFunc()
		for {
			line, err := term.ReadLine()
			if err != nil {
				log.Warning.Printf("Wait confirm user readLine exit: %s", err.Error())
				return
			}
			switch line {
			case "quit", "q":
				log.Info.Printf("Conn[%s] quit confirm", userConn.ID()[:8])
				return
			}
		}
	}()
	reviewers := srv.GetReviewers()
	titleMsg := "Need confirm to login"
	reviewersMsg := fmt.Sprintf("Ticket Reviewers: %s", strings.Join(reviewers, ", "))
	waitMsg := "Please waiting for the reviewers to confirm, enter q to exit. "
	common.IgnoreErrWriteString(userConn, titleMsg)
	common.IgnoreErrWriteString(userConn, common.CharNewLine)
	common.IgnoreErrWriteString(userConn, reviewersMsg)
	common.IgnoreErrWriteString(userConn, common.CharNewLine)
	common.IgnoreErrWriteString(userConn, common.CharNewLine)
	go func() {
		delay := 0
		for {
			select {
			case <-ctx.Done():
				return
			default:
				delayS := fmt.Sprintf("%ds", delay)
				data := strings.Repeat("\x08", len(delayS)+len(waitMsg)) + waitMsg + delayS
				common.IgnoreErrWriteString(userConn, data)
				time.Sleep(time.Second)
				delay += 1
			}
		}
	}()

	status := srv.WaitLoginConfirm(ctx)
	cancelFunc()
	processor := srv.GetApprover()
	var success bool
	statusMsg := "Unknown status"
	switch status {
	case auth.StatusApprove:
		formatMsg := "%s approved"
		statusMsg = common.WrapperString(fmt.Sprintf(formatMsg, processor), common.Green)
		success = true
	case auth.StatusReject:
		formatMsg := "%s rejected"
		statusMsg = common.WrapperString(fmt.Sprintf(formatMsg, processor), common.Red)
	case auth.StatusCancel:
		statusMsg = common.WrapperString("Cancel confirm", common.Red)
	}
	log.Info.Printf("Conn[%s] Login Confirm result: %s", userConn.ID()[:8], statusMsg)
	common.IgnoreErrWriteString(userConn, common.CharNewLine)
	common.IgnoreErrWriteString(userConn, statusMsg)
	common.IgnoreErrWriteString(userConn, common.CharNewLine)
	return success
}
