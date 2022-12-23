package srvconn

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/handewo/gojump/pkg/log"
	gossh "golang.org/x/crypto/ssh"
)

type UserSSHClient struct {
	ID   string // userID_assetID_systemUserID_systemUsername
	data map[*SSHClient]int64
	name string
}

func (u *UserSSHClient) AddClient(client *SSHClient) {
	u.data[client] = time.Now().UnixNano()
}

func (u *UserSSHClient) GetClient() *SSHClient {
	var selectClient *SSHClient
	var refCount int32
	// 取引用最少的 SSHClient
	for clientItem := range u.data {
		if refCount <= clientItem.RefCount() {
			refCount = clientItem.RefCount()
			selectClient = clientItem
		}
	}
	return selectClient
}

func (u *UserSSHClient) recycleClients() {
	needRemovedClients := make([]*SSHClient, 0, len(u.data))
	for client := range u.data {
		if client.RefCount() <= 0 {
			needRemovedClients = append(needRemovedClients, client)
			_ = client.Close()
		}
	}
	if len(needRemovedClients) > 0 {
		for i := range needRemovedClients {
			delete(u.data, needRemovedClients[i])
		}
		log.Info.Printf("Remove %d clients (%s) remain %d",
			len(needRemovedClients), u.name, len(u.data))
	}
}

func (u *UserSSHClient) Count() int {
	return len(u.data)
}

func newSSHManager() *SSHManager {
	m := SSHManager{
		storeChan:  make(chan *storeClient),
		reqChan:    make(chan string),
		resultChan: make(chan *SSHClient),
		searchChan: make(chan string),
	}
	go m.run()
	return &m
}

func AddClientCache(key string, client *SSHClient) {
	sshManager.AddClientCache(key, client)
}

type SSHManager struct {
	storeChan  chan *storeClient
	reqChan    chan string // reqId
	resultChan chan *SSHClient
	searchChan chan string // prefix
}

func (s *SSHManager) run() {
	tick := time.NewTicker(time.Minute)
	defer tick.Stop()
	data := make(map[string]*UserSSHClient)
	latestVisited := time.Now()
	for {
		select {
		case now := <-tick.C:
			/*
				1. 5 分钟无访问则 让所有的 UserSSHClient recycleClients
				2. 并清理 count==0 的 UserSSHClient
			*/
			if now.After(latestVisited.Add(time.Minute)) {
				needRemovedClients := make([]string, 0, len(data))
				for key, userClient := range data {
					userClient.recycleClients()
					if userClient.Count() == 0 {
						needRemovedClients = append(needRemovedClients, key)
					}
				}
				if len(needRemovedClients) > 0 {
					for i := range needRemovedClients {
						delete(data, needRemovedClients[i])
					}
					log.Info.Printf("Remove %d user clients remain %d",
						len(needRemovedClients), len(data))
				}
			}
			continue
		case reqKey := <-s.reqChan:
			var foundClient *SSHClient
			if userClient, ok := data[reqKey]; ok {
				foundClient = userClient.GetClient()
				log.Info.Printf("Found client(%s) and remain %d",
					foundClient, userClient.Count())
			}
			s.resultChan <- foundClient

		case prefixKey := <-s.searchChan:
			var foundClient *SSHClient
			for key, userClient := range data {
				if strings.HasPrefix(key, prefixKey) {
					foundClient = userClient.GetClient()
					log.Info.Printf("Found client(%s) and remain %d",
						foundClient, userClient.Count())
					break
				}
			}
			s.resultChan <- foundClient
		case reqClient := <-s.storeChan:
			userClient, ok := data[reqClient.reqId]
			if !ok {
				userClient = &UserSSHClient{
					ID:   reqClient.reqId,
					name: reqClient.SSHClient.String(),
					data: make(map[*SSHClient]int64),
				}
				data[reqClient.reqId] = userClient
			}
			userClient.AddClient(reqClient.SSHClient)
			log.Info.Printf("Store new client(%s) remain %d", reqClient.String(), userClient.Count())
		}

		latestVisited = time.Now()
	}
}

func (s *SSHManager) getClientFromCache(key string) (*SSHClient, bool) {
	s.reqChan <- key
	client := <-s.resultChan
	return client, client != nil
}

func (s *SSHManager) AddClientCache(key string, client *SSHClient) {
	s.storeChan <- &storeClient{
		reqId:     key,
		SSHClient: client,
	}
}

type storeClient struct {
	reqId string
	*SSHClient
}

type SSHClientOption func(conf *SSHClientOptions)

type SSHClientOptions struct {
	Host         string
	Port         string
	Username     string
	Password     string
	PrivateKey   string
	Passphrase   string
	Timeout      int
	keyboardAuth gossh.KeyboardInteractiveChallenge
	PrivateAuth  gossh.Signer
}

func (cfg *SSHClientOptions) AuthMethods() []gossh.AuthMethod {
	authMethods := make([]gossh.AuthMethod, 0, 3)

	if cfg.PrivateKey != "" {
		var (
			signer gossh.Signer
			err    error
		)
		if cfg.Passphrase != "" {
			// 先使用 passphrase 解析 PrivateKey
			if signer, err = gossh.ParsePrivateKeyWithPassphrase([]byte(cfg.PrivateKey),
				[]byte(cfg.Passphrase)); err == nil {
				authMethods = append(authMethods, gossh.PublicKeys(signer))
			}
		}
		if err != nil || cfg.Passphrase == "" {
			// 1. 如果之前使用解析失败，则去掉 passphrase，则尝试直接解析 PrivateKey 防止错误的passphrase
			// 2. 如果没有 Passphrase 则直接解析 PrivateKey
			if signer, err = gossh.ParsePrivateKey([]byte(cfg.PrivateKey)); err == nil {
				authMethods = append(authMethods, gossh.PublicKeys(signer))
			}
		}
	}
	if cfg.PrivateAuth != nil {
		authMethods = append(authMethods, gossh.PublicKeys(cfg.PrivateAuth))
	}
	if cfg.Password != "" {
		authMethods = append(authMethods, gossh.Password(cfg.Password))
	}
	if cfg.keyboardAuth != nil {
		authMethods = append(authMethods, gossh.KeyboardInteractive(cfg.keyboardAuth))
	}
	if cfg.keyboardAuth == nil && cfg.Password != "" {
		cfg.keyboardAuth = func(user, instruction string, questions []string, echos []bool) (answers []string, err error) {
			if len(questions) == 0 {
				return []string{}, nil
			}
			return []string{cfg.Password}, nil
		}
		authMethods = append(authMethods, gossh.KeyboardInteractive(cfg.keyboardAuth))
	}

	return authMethods
}

func SSHClientUsername(username string) SSHClientOption {
	return func(args *SSHClientOptions) {
		args.Username = username
	}
}

func SSHClientPassword(password string) SSHClientOption {
	return func(args *SSHClientOptions) {
		args.Password = password
	}
}

func SSHClientPrivateKey(privateKey string) SSHClientOption {
	return func(args *SSHClientOptions) {
		args.PrivateKey = privateKey
	}
}

func SSHClientPassphrase(passphrase string) SSHClientOption {
	return func(args *SSHClientOptions) {
		args.Passphrase = passphrase
	}
}

func SSHClientHost(host string) SSHClientOption {
	return func(args *SSHClientOptions) {
		args.Host = host
	}
}

func SSHClientPort(port int) SSHClientOption {
	return func(args *SSHClientOptions) {
		args.Port = strconv.Itoa(port)
	}
}

func SSHClientTimeout(timeout int) SSHClientOption {
	return func(args *SSHClientOptions) {
		args.Timeout = timeout
	}
}

func SSHClientPrivateAuth(privateAuth gossh.Signer) SSHClientOption {
	return func(args *SSHClientOptions) {
		args.PrivateAuth = privateAuth
	}
}

func SSHClientKeyboardAuth(keyboardAuth gossh.KeyboardInteractiveChallenge) SSHClientOption {
	return func(conf *SSHClientOptions) {
		conf.keyboardAuth = keyboardAuth
	}
}

func NewSSHClient(opts ...SSHClientOption) (*SSHClient, error) {
	cfg := &SSHClientOptions{
		Host: "127.0.0.1",
		Port: "22",
	}
	for _, setter := range opts {
		setter(cfg)
	}
	return NewSSHClientWithCfg(cfg)
}

var (
	ErrGatewayDial = errors.New("gateway dial addr failed")
	ErrSSHClient   = errors.New("new ssh client failed")
)

func NewSSHClientWithCfg(cfg *SSHClientOptions) (*SSHClient, error) {
	gosshCfg := gossh.ClientConfig{
		User:            cfg.Username,
		Auth:            cfg.AuthMethods(),
		Timeout:         time.Duration(cfg.Timeout) * time.Second,
		HostKeyCallback: gossh.InsecureIgnoreHostKey(),
		Config:          createSSHConfig(),
	}
	destAddr := net.JoinHostPort(cfg.Host, cfg.Port)
	gosshClient, err := gossh.Dial("tcp", destAddr, &gosshCfg)
	if err != nil {
		return nil, err
	}
	return &SSHClient{Client: gosshClient, Cfg: cfg,
		traceSessionMap: make(map[*gossh.Session]time.Time)}, nil
}

type SSHClient struct {
	*gossh.Client
	Cfg         *SSHClientOptions
	ProxyClient *SSHClient

	sync.Mutex

	traceSessionMap map[*gossh.Session]time.Time

	refCount int32
}

func (s *SSHClient) String() string {
	return fmt.Sprintf("%s@%s:%s", s.Cfg.Username,
		s.Cfg.Host, s.Cfg.Port)
}

func (s *SSHClient) Close() error {
	if s.ProxyClient != nil {
		_ = s.ProxyClient.Close()
		log.Info.Printf("SSHClient(%s) proxy (%s) close", s, s.ProxyClient)
	}
	err := s.Client.Close()
	log.Info.Printf("SSHClient(%s) close", s)
	return err
}

func (s *SSHClient) RefCount() int32 {
	return atomic.LoadInt32(&s.refCount)
}

func (s *SSHClient) AcquireSession() (*gossh.Session, error) {
	atomic.AddInt32(&s.refCount, 1)
	sess, err := s.Client.NewSession()
	if err != nil {
		atomic.AddInt32(&s.refCount, -1)
		return nil, err
	}
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	s.traceSessionMap[sess] = time.Now()
	log.Info.Printf("SSHClient(%s) session add one ", s)
	return sess, nil
}

func (s *SSHClient) ReleaseSession(sess *gossh.Session) {
	atomic.AddInt32(&s.refCount, -1)
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	delete(s.traceSessionMap, sess)
	log.Info.Printf("SSHClient(%s) release one session remain %d", s, len(s.traceSessionMap))
}

func createSSHConfig() gossh.Config {
	var cfg gossh.Config
	cfg.SetDefaults()
	cfg.Ciphers = append(cfg.Ciphers, notRecommendCiphers...)
	cfg.KeyExchanges = append(cfg.KeyExchanges, notRecommendKeyExchanges...)
	return cfg
}

var (
	notRecommendCiphers = []string{
		"arcfour256", "arcfour128", "arcfour",
		"aes128-cbc", "3des-cbc",
	}

	notRecommendKeyExchanges = []string{
		"diffie-hellman-group1-sha1", "diffie-hellman-group-exchange-sha1",
		"diffie-hellman-group-exchange-sha256",
	}
)
