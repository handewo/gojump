package proxy

import (
	"context"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/handewo/gojump/pkg/common"
	"github.com/handewo/gojump/pkg/log"
	"github.com/handewo/gojump/pkg/srvconn"
)

type SwitchSession struct {
	ID string

	MaxIdleTime   int
	keepAliveTime int

	ctx    context.Context
	cancel context.CancelFunc

	p *Server

	terminateAdmin atomic.Value // 终断会话的管理员名称
}

func (s *SwitchSession) Terminate(username string) {
	select {
	case <-s.ctx.Done():
		return
	default:
		s.setTerminateAdmin(username)
	}
	s.cancel()
	log.Info.Printf("Session[%s] receive terminate task from admin %s", s.ID[:8], username)
}

func (s *SwitchSession) setTerminateAdmin(username string) {
	s.terminateAdmin.Store(username)
}

func (s *SwitchSession) loadTerminateAdmin() string {
	return s.terminateAdmin.Load().(string)
}

func (s *SwitchSession) SessionID() string {
	return s.ID
}

// Bridge 桥接两个链接
func (s *SwitchSession) Bridge(userConn UserConnection, srvConn srvconn.ServerConnection) (err error) {

	done := make(chan struct{})

	srvChan := make(chan []byte, 1)
	userChan := make(chan []byte, 1)

	replayRecorder := s.p.GetReplayRecorder()

	defer func() {
		close(done)
		_ = userConn.Close()
		_ = srvConn.Close()
		replayRecorder.End()
	}()

	winCh := userConn.WinCh()
	maxIdleTime := time.Duration(s.MaxIdleTime) * time.Minute
	lastActiveTime := time.Now()
	tick := time.NewTicker(30 * time.Second)
	defer tick.Stop()

	exitSignal := make(chan struct{}, 1)
	go func() {
		var (
			exitFlag bool
			err      error
			nr       int
		)
		for {
			buf := make([]byte, 1024)
			nr, err = srvConn.Read(buf)
			if nr > 0 {
				select {
				case srvChan <- buf[:nr]:
				case <-done:
					exitFlag = true
					log.Info.Printf("Session[%s] done", s.ID[:8])
				}
				if exitFlag {
					break
				}
			}
			if err != nil {
				if err != io.EOF {
					log.Error.Printf("Session[%s] srv read err: %s", s.ID[:8], err)
				}
				break
			}
		}
		log.Debug.Printf("Session[%s] srv read end", s.ID[:8])
		exitSignal <- struct{}{}
		close(srvChan)
	}()

	go func() {
		for {
			buf := make([]byte, 1024)
			nr, err := userConn.Read(buf)
			if nr > 0 {
				userChan <- buf[:nr]
			}
			if err != nil {
				log.Warning.Printf("Session[%s] user read err: %s", s.ID[:8], err)
				break
			}
		}
		log.Debug.Printf("Session[%s] user read end", s.ID[:8])
		exitSignal <- struct{}{}
	}()

	keepAliveTime := time.Duration(s.keepAliveTime) * time.Second
	keepAliveTick := time.NewTicker(keepAliveTime)
	defer keepAliveTick.Stop()
	for {
		select {
		// 检测是否超过最大空闲时间
		case now := <-tick.C:
			outTime := lastActiveTime.Add(maxIdleTime)
			if now.After(outTime) {
				log.Info.Printf("Session[%s] idle more than %d minutes, disconnect", s.ID[:8], s.MaxIdleTime)
				return
			}
			if s.p.CheckPermissionExpired(now) {
				log.Info.Printf("Session[%s] permission has expired, disconnect", s.ID[:8])
				return
			}
			continue
			// 手动结束
		case <-s.ctx.Done():
			adminUser := s.loadTerminateAdmin()
			msg := fmt.Sprintf("Terminated by admin %s", adminUser)
			msg = common.WrapperWarn(msg)
			log.Info.Printf("Session[%s]: %s", s.ID[:8], msg)
			return
			// 监控窗口大小变化
		case win, ok := <-winCh:
			if !ok {
				return
			}
			_ = srvConn.SetWinSize(win.Width, win.Height)
			log.Debug.Printf("Session[%s] Window server change: %d*%d",
				s.ID[:8], win.Width, win.Height)
			// 经过parse处理的server数据，发给user
		case p, ok := <-srvChan:
			if !ok {
				return
			}
			replayRecorder.Record(p)
			if _, err := userConn.Write(p); err != nil {
				log.Error.Printf("Session[%s] userConn write err: %s", s.ID[:8], err)
			}
			// 经过parse处理的user数据，发给server
		case p, ok := <-userChan:
			if !ok {
				return
			}
			if _, err := srvConn.Write(p); err != nil {
				log.Error.Printf("Session[%s] srvConn write err: %s", s.ID[:8], err)
			}

		case now := <-keepAliveTick.C:
			if now.After(lastActiveTime.Add(keepAliveTime)) {
				if err := srvConn.KeepAlive(); err != nil {
					log.Error.Printf("Session[%s] srvCon keep alive err: %s", s.ID[:8], err)
				}
			}
			continue
		case <-userConn.Context().Done():
			log.Info.Printf("Session[%s]: user conn context done", s.ID[:8])
			return
		case <-exitSignal:
			log.Debug.Printf("Session[%s] end by exit signal", s.ID[:8])
			return
		}
		lastActiveTime = time.Now()
	}
}

var sessManager = newSessionManager()

func GetSessionById(id string) (s *SwitchSession, ok bool) {
	s, ok = sessManager.Get(id)
	return
}

func GetAliveSessions() []string {
	return sessManager.Range()
}

func AddCommonSwitch(s *SwitchSession) {
	sessManager.Add(s.ID, s)
}

func RemoveCommonSwitch(s *SwitchSession) {
	sessManager.Delete(s.ID)
}

func newSessionManager() *sessionManager {
	return &sessionManager{
		data: make(map[string]*SwitchSession),
	}
}

type sessionManager struct {
	data map[string]*SwitchSession
	sync.Mutex
}

func (s *sessionManager) Add(id string, sess *SwitchSession) {
	s.Lock()
	defer s.Unlock()
	s.data[id] = sess
}
func (s *sessionManager) Get(id string) (sess *SwitchSession, ok bool) {
	s.Lock()
	defer s.Unlock()
	sess, ok = s.data[id]
	return
}

func (s *sessionManager) Delete(id string) {
	s.Lock()
	defer s.Unlock()
	delete(s.data, id)
}

func (s *sessionManager) Range() []string {
	sids := make([]string, 0, len(s.data))
	s.Lock()
	defer s.Unlock()
	for sid := range s.data {
		sids = append(sids, sid)
	}
	return sids
}
