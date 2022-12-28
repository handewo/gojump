package core

import (
	"fmt"
	"time"

	"github.com/genjidb/genji/document"
	"github.com/genjidb/genji/types"
	"github.com/handewo/gojump/pkg/common"
	"github.com/handewo/gojump/pkg/log"
	"github.com/handewo/gojump/pkg/model"
)

func (c *Core) QueryUserLog() ([]string, error) {
	log := make([]string, 0, 10)
	res, err := c.db.Query("SELECT datetime,type,user,log FROM USERLOG")
	if err != nil {
		return nil, err
	}
	defer res.Close()
	err = res.Iterate(func(d types.Document) error {
		var dt, t, u, l string
		err = document.Scan(d, &dt, &t, &u, &l)
		if err != nil {
			return err
		}
		log = append(log, fmt.Sprintf("%s|%10s|%10s|%s", dt, t, u, l))
		return nil
	})
	if err != nil {
		return nil, err
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
	err := InsertData(c.db, model.InsertUserLog, &lg)
	if err != nil {
		log.Error.Printf("insert authentication log failed, %s", err)
	}
}
