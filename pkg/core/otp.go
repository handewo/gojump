package core

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/handewo/gojump/pkg/log"
)

func (c *Core) GenOTPassword(name string) string {
	c.otpLock.Lock()
	pass, ok := c.otpassword[name]
	if ok {
		c.otpLock.Unlock()
		return pass
	}
	pass = randomPass()
	c.otpassword[name] = pass
	c.otpLock.Unlock()

	go c.clearOTPassword(name, 90)
	return pass
}

func (c *Core) clearOTPassword(name string, after int64) {
	time.Sleep(time.Duration(after) * time.Second)
	c.otpLock.Lock()
	defer c.otpLock.Unlock()
	_, ok := c.otpassword[name]
	if !ok {
		return
	}
	delete(c.otpassword, name)
}

func (c *Core) verifyOTP(name string, inputPass string) bool {
	if inputPass == "" {
		return false
	}
	c.otpLock.Lock()
	pass, ok := c.otpassword[name]
	c.otpLock.Unlock()
	if !ok {
		return false
	}
	res := pass == inputPass
	if res {
		c.clearOTPassword(name, 0)
	}
	return res

}
func randomPass() string {
	n, err := rand.Int(rand.Reader, big.NewInt(99999999))
	if err == nil {
		return fmt.Sprintf("%08v", n.Uint64())
	}
	log.Error.Printf("generate random number error: %s", err)
	return "10957890"
}
