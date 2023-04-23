package core

import (
	"errors"
	"fmt"
	"time"

	"github.com/handewo/gojump/pkg/common"
	"github.com/handewo/gojump/pkg/log"
	"github.com/handewo/gojump/pkg/model"
)

func (c *Core) QueryUserLog() ([]string, error) {
	v, err := c.db.QueryStructs(model.UserlogType, "SELECT * FROM USERLOG")
	if err != nil {
		return nil, err
	}

	logs, ok := v.([]model.UserLog)
	if !ok {
		return nil, errors.New("invalid value type")
	}
	if err != nil {
		return nil, err
	}

	log := make([]string, 0, 10)
	for _, l := range logs {
		log = append(log, fmt.Sprintf("%s|%10s|%10s|%s", l.Datetime, l.Type, l.User, l.Log))
	}
	return log, nil
}

func (c *Core) InteractiveLog(user string) {
	date := time.Now().Format(common.LogFormat)
	lg := model.UserLog{
		Datetime: date,
		Type:     "term",
		User:     user,
		Log:      "exit terminal",
	}
	err := c.db.InsertData("INSERT INTO USERLOG VALUES ?", &lg)
	if err != nil {
		log.Error.Printf("insert authentication log failed, %s", err)
	}
}

func (c *Core) InsertLog(tp, user, msg string) {
	date := time.Now().Format(common.LogFormat)
	lg := model.UserLog{
		Datetime: date,
		Type:     tp,
		User:     user,
		Log:      msg,
	}
	err := c.db.InsertData("INSERT INTO USERLOG VALUES ?", &lg)
	if err != nil {
		log.Error.Printf("insert log failed, %s", err)
	}
}
