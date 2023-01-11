package core

import (
	"fmt"
	"os"
	"sync"

	"github.com/handewo/gojump/pkg/config"
	"github.com/handewo/gojump/pkg/log"
	"github.com/handewo/gojump/pkg/model"
)

type Core struct {
	db            DB
	sessLock      sync.RWMutex
	session       map[string]model.Session
	otpLock       sync.Mutex
	otpassword    map[string]string
	loginLock     sync.RWMutex
	tryLoginCount map[string]uint64
}

func NewCore() *Core {

	path := config.GlobalConfig.GenjiDbPath
	_, err := os.Stat(path)
	if err != nil {
		log.Fatal.Fatal(err)
	}

	db, err := newDataBase(config.GlobalConfig.Database)
	if err != nil {
		log.Fatal.Fatal(err)
	}
	session := make(map[string]model.Session, 100)
	otpass := make(map[string]string, 4)
	tryLoginCnt := make(map[string]uint64, 10)
	return &Core{
		db:            db,
		session:       session,
		otpassword:    otpass,
		tryLoginCount: tryLoginCnt,
	}

}

func (c *Core) Close() {
	err := c.db.Close()
	if err != nil {
		log.Error.Print(err)
		return
	}
	log.Debug.Print("database closed")
}

func newDataBase(db string) (DB, error) {
	switch db {
	case "genji":
		return NewGenji(config.GlobalConfig.GenjiDbPath)
	}
	return nil, fmt.Errorf("%s is unsupported", db)
}
