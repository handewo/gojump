package model

import "fmt"

// USER TABLE
type User struct {
	ID            string `json:"id"`
	Username      string `json:"name"`
	Role          string `json:"role"`
	ExpiredAt     int64
	OTPLevel      int  `json:"otp_level"`
	IsActive      bool `json:"is_active"`
	NodeIDs       []string
	AddrWhiteList []string
}

func (u *User) String() string {
	return fmt.Sprint(u.Username)
}

type UserSecret struct {
	UserID         string
	Password       string
	PrivateKey     string
	AuthorizedKeys []string
}
