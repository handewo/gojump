package handler

import (
	"fmt"
	"html/template"
	"io"

	"github.com/handewo/gojump/pkg/common"
	"github.com/handewo/gojump/pkg/log"
	"github.com/handewo/gojump/pkg/model"
)

type MenuItem struct {
	id       int
	instruct string
	helpText string
}

type Menu []MenuItem

type ColorMeta struct {
	GreenBoldColor string
	ColorEnd       string
}

func (h *InteractiveHandler) displayBanner(sess io.ReadWriter, termConf *model.TerminalConfig) {
	defaultTitle := common.WrapperTitle("Welcome to use GOJump")
	menu := Menu{
		{id: 1, instruct: "part IP, Hostname, Comment", helpText: "to search login if unique"},
		{id: 2, instruct: "/ + IP, Hostname, Comment", helpText: "to search, such as: /192.168"},
		{id: 3, instruct: "p", helpText: "display the host you have permission"},
		{id: 4, instruct: "g", helpText: "display the node that you have permission"},
		{id: 5, instruct: "r", helpText: "refresh your assets and nodes"},
		{id: 6, instruct: "h", helpText: "print help"},
		{id: 7, instruct: "q", helpText: "exit"},
	}

	title := defaultTitle
	if termConf.HeaderTitle != "" {
		title = termConf.HeaderTitle
	}

	prefix := common.CharClear + common.CharTab + common.CharTab + common.CharTab
	suffix := common.CharNewLine + common.CharNewLine
	welcomeMsg := prefix + title + suffix
	_, err := io.WriteString(sess, welcomeMsg)
	if err != nil {
		log.Error.Printf("Send to client error, %s", err)
		return
	}
	cm := ColorMeta{GreenBoldColor: "\033[1;32m", ColorEnd: "\033[0m"}
	for _, v := range menu {
		line := fmt.Sprintf("\t%d) Enter {{.GreenBoldColor}}%s{{.ColorEnd}} to %s.%s",
			v.id, v.instruct, v.helpText, "\r\n")
		tmpl := template.Must(template.New("item").Parse(line))
		if err := tmpl.Execute(sess, cm); err != nil {
			log.Error.Print(err)
		}
	}
}

func displayAdminHelp(sess io.ReadWriter) {
	title := common.WrapperTitle("GOJump Admin")
	menu := Menu{
		{id: 1, instruct: "otp USERNAME", helpText: "generate otp for user"},
		{id: 2, instruct: "list TABLE", helpText: "list [USERLOG, TICKET]"},
		{id: 3, instruct: "ticket", helpText: "list pending tickets"},
		{id: 4, instruct: "approve TICKETID", helpText: "approve the ticket"},
		{id: 5, instruct: "reject TICKETID", helpText: "reject the ticket"},
		{id: 6, instruct: "h", helpText: "print help"},
		{id: 7, instruct: "q", helpText: "exit"},
	}

	prefix := common.CharClear + common.CharTab + common.CharTab + common.CharTab
	suffix := common.CharNewLine + common.CharNewLine
	_, err := io.WriteString(sess, prefix+title+suffix)
	if err != nil {
		log.Error.Printf("Send to client error, %s", err)
		return
	}
	cm := ColorMeta{GreenBoldColor: "\033[1;32m", ColorEnd: "\033[0m"}
	for _, v := range menu {
		line := fmt.Sprintf("\t%d) Enter {{.GreenBoldColor}}%s{{.ColorEnd}} to %s.%s",
			v.id, v.instruct, v.helpText, "\r\n")
		tmpl := template.Must(template.New("item").Parse(line))
		if err := tmpl.Execute(sess, cm); err != nil {
			log.Error.Print(err)
		}
	}
}
