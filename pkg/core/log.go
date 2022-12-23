package core

import (
	"fmt"

	"github.com/genjidb/genji/document"
	"github.com/genjidb/genji/types"
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
