package core

import (
	"fmt"
	"time"

	"github.com/handewo/gojump/pkg/common"
	"github.com/handewo/gojump/pkg/model"
)

func (c *Core) CreateSession(session model.Session) error {
	c.sessLock.Lock()
	defer c.sessLock.Unlock()
	c.session[session.ID] = session
	return nil
}

func (c *Core) SessionSuccess(id string) error {
	c.sessLock.RLock()
	s, ok := c.session[id]
	c.sessLock.RUnlock()
	if !ok {
		return nil
	}
	date := time.Now().Format(common.LogFormat)
	log := model.UserLog{
		Datetime: date,
		Type:     "session",
		User:     s.User,
		Log:      fmt.Sprintf("login to %s successfully", s.Asset),
	}
	err := InsertData(c.db, model.InsertUserLog, &log)
	return err
}

func (c *Core) SessionFailed(id string, cause error) error {
	c.sessLock.RLock()
	s, ok := c.session[id]
	c.sessLock.RUnlock()
	if !ok {
		return nil
	}
	date := time.Now().Format(common.LogFormat)
	log := model.UserLog{
		Datetime: date,
		Type:     "session",
		User:     s.User,
		Log:      fmt.Sprintf("login to %s failed, error: %s", s.Asset, cause),
	}
	err := InsertData(c.db, model.InsertUserLog, &log)
	c.sessLock.Lock()
	delete(c.session, id)
	c.sessLock.Unlock()
	return err
}

func (c *Core) SessionDisconnect(id string) error {
	c.sessLock.RLock()
	s, ok := c.session[id]
	c.sessLock.RUnlock()
	if !ok {
		return nil
	}
	date := time.Now().Format(common.LogFormat)
	log := model.UserLog{
		Datetime: date,
		Type:     "session",
		User:     s.User,
		Log:      fmt.Sprintf("disconnected to %s", s.Asset),
	}
	err := InsertData(c.db, model.InsertUserLog, &log)
	c.sessLock.Lock()
	delete(c.session, id)
	c.sessLock.Unlock()
	return err
}
