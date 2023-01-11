package core

import "github.com/handewo/gojump/pkg/model"

type DB interface {
	Close() error
	GetTerminalConfig() (model.TerminalConfig, error)
	QueryStruct(st interface{}, sql string, cond ...interface{}) error
	QueryStructs(mod int, sql string, cond ...interface{}) (interface{}, error)
	QueryOneField(f interface{}, sql string, cond ...interface{}) error
	QueryTwoField(f1 interface{}, f2 interface{}, sql string, cond ...interface{}) error
	QueryOneFieldMutilRows(sql string, cond ...interface{}) ([]string, error)
	InsertData(sql string, data ...interface{}) error
	UpdateData(sql string, args ...interface{}) error
}
