package model

import "time"

type Session struct {
	ID           string    `json:"id"`
	User         string    `json:"user"`
	Asset        string    `json:"asset"`
	SystemUser   string    `json:"system_user"`
	LoginFrom    string    `json:"login_from"`
	RemoteAddr   string    `json:"remote_addr"`
	Protocol     string    `json:"protocol"`
	DateStart    time.Time `json:"date_start"`
	UserID       string    `json:"user_id"`
	AssetID      string    `json:"asset_id"`
	SystemUserID string    `json:"system_user_id"`
}
