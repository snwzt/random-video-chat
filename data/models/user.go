package models

type User struct {
	Username string `json:"username"`
	IPAddr   string `json:"ip_addr"`
}
