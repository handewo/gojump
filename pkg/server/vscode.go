package server

import (
	"fmt"

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
	msg := fmt.Sprintf("vscode connect to %s successfully", vsReq.client)
	s.core.InsertLog("vscode", vsReq.user.Username, msg)
	s.vscodeClients[vsReq.reqId] = vsReq
}

func (s *server) deleteVSCodeReq(vsReq *vscodeReq) {
	s.Lock()
	defer s.Unlock()
	msg := fmt.Sprintf("vscode disconnect to %s", vsReq.client)
	s.core.InsertLog("vscode", vsReq.user.Username, msg)
	delete(s.vscodeClients, vsReq.reqId)
}
