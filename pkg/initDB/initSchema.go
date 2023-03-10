package initDB

import (
	"fmt"

	"github.com/genjidb/genji"
	"github.com/handewo/gojump/pkg/log"
)

func InitSchema(db *genji.DB) {
	schemas := []string{"TERMINALCONF", "USER", "ASSET", "NODE",
		"USERSECRET", "SYSTEMUSER", "ASSETUSERINFO", "USERLOG", "LOGINTICKET"}

	var err error
	for _, v := range schemas {
		err = db.Exec(fmt.Sprintf("CREATE TABLE %s", v))
		if err != nil {
			log.Fatal.Print(err)
		}
	}
}
