package core

import (
	"errors"

	"github.com/genjidb/genji"
	"github.com/genjidb/genji/document"
	"github.com/genjidb/genji/types"
	"github.com/handewo/gojump/pkg/model"
)

type Genji struct {
	db *genji.DB
}

type Model interface {
	model.Asset | model.Node | model.User | model.SystemUser | model.AssetUserInfo | model.UserLog | model.LoginTicket | model.UserSecret
}

func NewGenji(path string) (DB, error) {
	db, err := genji.Open(path)
	if err != nil {
		return nil, err
	}
	return &Genji{db: db}, nil
}

func (g *Genji) Close() error {
	return g.db.Close()
}

func (g *Genji) GetTerminalConfig() (model.TerminalConfig, error) {
	res, err := g.db.Query("SELECT * FROM TERMINALCONF")
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

func (g *Genji) QueryOneField(f interface{}, sql string, cond ...interface{}) error {
	res, err := g.db.Query(sql, cond...)
	if err != nil {
		return err
	}
	defer res.Close()
	err = res.Iterate(func(d types.Document) error {
		err = document.Scan(d, f)
		return err
	})
	return err
}

func (g *Genji) QueryTwoField(f1 interface{}, f2 interface{}, sql string, cond ...interface{}) error {
	res, err := g.db.Query(sql, cond...)
	if err != nil {
		return err
	}
	defer res.Close()
	err = res.Iterate(func(d types.Document) error {
		err = document.Scan(d, f1, f2)
		return err
	})
	return err
}

func (g *Genji) QueryStruct(st interface{}, sql string, cond ...interface{}) error {
	res, err := g.db.Query(sql, cond...)
	if err != nil {
		return err
	}
	defer res.Close()
	err = res.Iterate(func(d types.Document) error {
		err = document.StructScan(d, st)
		return err
	})
	return err
}

func (g *Genji) InsertData(sql string, data ...interface{}) error {
	err := g.db.Exec(sql, data...)
	return err
}

func (g *Genji) QueryStructs(mod int, sql string, cond ...interface{}) (interface{}, error) {
	switch mod {
	case model.AssetType:
		return queryStructs[model.Asset](g.db, sql, cond...)
	case model.NodeType:
		return queryStructs[model.Node](g.db, sql, cond...)
	case model.UserType:
		return queryStructs[model.User](g.db, sql, cond...)
	case model.SystemUserType:
		return queryStructs[model.SystemUser](g.db, sql, cond...)
	case model.AssetUserInfoType:
		return queryStructs[model.AssetUserInfo](g.db, sql, cond...)
	case model.UserlogType:
		return queryStructs[model.UserLog](g.db, sql, cond...)
	case model.LoginTicketType:
		return queryStructs[model.LoginTicket](g.db, sql, cond...)
	case model.UserSecretType:
		return queryStructs[model.UserSecret](g.db, sql, cond...)
	}
	return nil, errors.New("invalid model type")
}

func queryStructs[M Model](db *genji.DB, sql string, cond ...interface{}) ([]M, error) {
	res, err := db.Query(sql, cond...)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	models := make([]M, 0, 10)
	err = res.Iterate(func(d types.Document) error {
		var m M
		err = document.StructScan(d, &m)
		if err != nil {
			return err
		}
		models = append(models, m)
		return nil
	})
	return models, err
}

func (g *Genji) QueryOneFieldMutilRows(sql string, cond ...interface{}) ([]string, error) {
	res, err := g.db.Query(sql, cond...)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	s := make([]string, 0, 5)
	err = res.Iterate(func(d types.Document) error {
		var r string
		err = document.Scan(d, &r)
		if err != nil {
			return err
		}
		s = append(s, r)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (g *Genji) UpdateData(sql string, args ...interface{}) error {
	err := g.db.Exec(sql, args...)
	return err
}
