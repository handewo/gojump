package initDB

import (
	"github.com/genjidb/genji"
	"github.com/handewo/gojump/pkg/common"
	"github.com/handewo/gojump/pkg/log"
	"github.com/handewo/gojump/pkg/model"
)

func InitData(db *genji.DB) {
	key := common.GenerateEd25519Pem()

	d1 := model.TerminalConfig{
		AssetListPageSize:  "",
		AssetListSortByIp:  false,
		HeaderTitle:        "",
		PasswordAuth:       true,
		PublicKeyAuth:      true,
		MaxIdleTime:        90,
		HostKey:            key,
		EnableSessionShare: false,
	}
	err := InsertData(db, "INSERT INTO TERMINALCONF VALUES ?", &d1)
	if err != nil {
		log.Fatal.Print(err)
	}
}

func InsertData(db *genji.DB, sql string, data ...interface{}) error {
	err := db.Exec(sql, data...)
	return err
}
