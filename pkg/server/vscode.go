package server

import (
	"github.com/handewo/gojump/pkg/model"
	"github.com/handewo/gojump/pkg/srvconn"
)

type vscodeReq struct {
	reqId  string
	user   *model.User
	client *srvconn.SSHClient

	expireInfo *model.ExpireInfo
}

func (s *server) getVSCodeReq(reqId string) *vscodeReq {
	s.Lock()
	defer s.Unlock()
	return s.vscodeClients[reqId]
}

func (s *server) addVSCodeReq(vsReq *vscodeReq) {
	s.Lock()
	defer s.Unlock()
	s.vscodeClients[vsReq.reqId] = vsReq
}

func (s *server) deleteVSCodeReq(vsReq *vscodeReq) {
	s.Lock()
	defer s.Unlock()
	delete(s.vscodeClients, vsReq.reqId)
}
