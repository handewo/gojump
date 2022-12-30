package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/handewo/gojump/pkg/common"
	"github.com/handewo/gojump/pkg/config"
	"github.com/handewo/gojump/pkg/log"
	"github.com/handewo/gojump/pkg/model"
)

func (h *InteractiveHandler) AdminSystem() {
	defer log.Info.Printf("Request %s: Admin %s stop interactive", h.sess.ID()[:8], h.user.Username)
	checkChan := make(chan bool)
	go h.checkMaxIdleTime(checkChan)
	displayAdminHelp(h.sess)
	for {
		checkChan <- true
		line, err := h.term.ReadLine()
		if err != nil {
			log.Debug.Printf("User %s close connect %s", h.user.Username, err)
			break
		}
		checkChan <- false
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		switch line[0] {
		case 'h':
			displayAdminHelp(h.sess)
			continue
		case 'q':
			return
		}
		words := strings.Split(line, " ")
		switch words[0] {
		case "list":
			h.listTable(words[1])
			continue
		case "ticket":
			h.listTable("PENDINGTICKET")
			continue
		case "approve":
			h.updateTicketState(words[1], model.TicketApproved)
			continue
		case "reject":
			h.updateTicketState(words[1], model.TicketRejected)
			continue
		case "otp":
			pass := h.core.GenOTPassword(words[1])
			msg := pass + common.CharNewLine
			h.term.Write([]byte(msg))
			continue
		case "help":
			displayAdminHelp(h.sess)
			continue
		case "exit", "quit":
			return
		}
	}
}

func (h *InteractiveHandler) updateTicketState(id string, state string) {
	err := h.core.UpdateTicketState(id, state, h.user.Username)
	if err != nil {
		log.Error.Printf("update ticket's state falied, %s", err)
		msg := common.WrapperString("Error", common.Red)
		common.IgnoreErrWriteString(h.sess, msg+common.CharNewLine)
		return
	}
	msg := common.WrapperString("Submit", common.Green)
	common.IgnoreErrWriteString(h.sess, msg+common.CharNewLine)
}

func (h *InteractiveHandler) listTable(table string) {
	var rows []string
	var title string
	var err error
	switch strings.ToUpper(table) {
	case "USERLOG":
		rows, err = h.core.QueryUserLog()
		if err != nil {
			log.Error.Printf("query error from USERLOG, %s", err)
			return
		}
		title = "           Date          |    Type  |    User  |    Log"
	case "TICKET":
		rows, err = h.core.QueryLoginTicket()
		if err != nil {
			log.Error.Printf("query error from LOGINTICKET, %s", err)
			return
		}
		title = "                 Ticket ID                | Application Date  |    User  |   Asset  |  SysUser |  Approver|   Approve Date    |   State"
	case "PENDINGTICKET":
		rows, err = h.core.QueryPengdingLoginTicket()
		if err != nil {
			log.Error.Printf("query error from LOGINTICKET, %s", err)
			return
		}
		title = "                 Ticket ID                | Application Date  |    User  |   Asset  |  SysUser "
	case "USER":
		rows, err = h.core.QueryAllUser()
		if err != nil {
			log.Error.Printf("query error from USER, %s", err)
			return
		}
		title = "        ID|    User  |  Role|      Expire At    |OTP Level|Active|      Nodes     |    White list"
	case "SYSUSER":
		rows, err = h.core.QueryAllSystemUser()
		if err != nil {
			log.Error.Printf("query error from SYSTEMUSER, %s", err)
			return
		}
		title = "        ID|    User  |Priority|Protocol|Password|Private Key|Comment"
	case "ASSET":
		rows, err = h.core.QueryAllAsset()
		if err != nil {
			log.Error.Printf("query error from ASSET, %s", err)
			return
		}
		title = "        ID|Asset Name| Hostname |                   IP                  |    OS    |Active| Platform |   Protocols   |Comment"
	case "NODE":
		rows, err = h.core.QueryAllNode()
		if err != nil {
			log.Error.Printf("query error from NODE, %s", err)
			return
		}
		title = "        ID|   Name   |    Key   |      Asset IDs"
	case "ASSETUSER":
		rows, err = h.core.QueryAssetUserInfo()
		if err != nil {
			log.Error.Printf("query error from ASSETUSERINFO, %s", err)
			return
		}
		title = "        ID|User ID|Asset ID|      Expire At    |Need Confirm|System User IDs"
	case "CONFIG":
		cfg, err := json.MarshalIndent(config.GlobalConfig, "", "    ")
		if err != nil {
			log.Error.Printf("json marshal error: %s", err)
			return
		}
		tcfg, err := json.MarshalIndent(h.terminalConf, "", "    ")
		if err != nil {
			log.Error.Printf("json marshal error: %s", err)
			return
		}
		_, err = io.WriteString(h.sess, string(cfg)+","+string(tcfg)+common.CharNewLine)
		if err != nil {
			log.Error.Printf("Send to client error, %s", err)
			return
		}
	}
	common.IgnoreErrWriteString(h.sess, title+common.CharNewLine)
	for i, v := range rows {
		l := fmt.Sprintf("%4d. %s", i+1, v)
		_, err := io.WriteString(h.sess, l+common.CharNewLine)
		if err != nil {
			log.Error.Printf("Send to client error, %s", err)
			return
		}
	}
}
