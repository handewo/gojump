package core

import (
	"os"
	"sync"

	"github.com/genjidb/genji"
	"github.com/genjidb/genji/document"
	"github.com/genjidb/genji/types"
	"github.com/handewo/gojump/pkg/config"
	"github.com/handewo/gojump/pkg/log"
	"github.com/handewo/gojump/pkg/model"
)

type Core struct {
	db         *genji.DB
	sessLock   sync.RWMutex
	session    map[string]model.Session
	otpLock    sync.Mutex
	otpassword map[string]string
}

func NewCore() *Core {

	path := config.GlobalConfig.DbFile
	_, err := os.Stat(path)
	if err != nil {
		log.Fatal.Fatal(err)
	}

	db, err := genji.Open(path)
	if err != nil {
		log.Fatal.Fatal(err)
	}

	session := make(map[string]model.Session, 100)
	otpass := make(map[string]string, 4)
	return &Core{
		db:         db,
		session:    session,
		otpassword: otpass,
	}

}

func (c *Core) Close() {
	c.db.Close()
}

func (c *Core) GetTerminalConfig() (conf model.TerminalConfig, err error) {
	res, err := c.db.Query("SELECT * FROM TERMINALCONF")
	if err != nil {
		return model.TerminalConfig{}, err
	}
	defer res.Close()
	t := model.TerminalConfig{}
	err = res.Iterate(func(d types.Document) error {
		err := document.StructScan(d, &t)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return model.TerminalConfig{}, err
	}
	return t, nil
}
