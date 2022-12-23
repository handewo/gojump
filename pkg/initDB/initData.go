package initDB

import (
	"github.com/genjidb/genji"
	"github.com/handewo/gojump/pkg/common"
	"github.com/handewo/gojump/pkg/core"
	"github.com/handewo/gojump/pkg/log"
	"github.com/handewo/gojump/pkg/model"
)

func InitData(db *genji.DB) {
	key := common.GenerateEd25519Pem()

	d1 := model.TerminalConfig{
		AssetListPageSize:   "",
		AssetListSortByIp:   false,
		HeaderTitle:         "",
		PasswordAuth:        true,
		PublicKeyAuth:       true,
		SessionKeepDuration: 90,
		MaxIdleTime:         20,
		HeartbeatDuration:   30,
		HostKey:             key,
		EnableSessionShare:  false,
	}
	err := core.InsertData(db, "INSERT INTO TERMINALCONF VALUES ?", &d1)
	if err != nil {
		log.Fatal.Print(err)
	}
}
