package model

type UserLog struct {
	Datetime string
	Type     string
	User     string
	Log      string
}

var InsertUserLog = "INSERT INTO USERLOG VALUES ?"
