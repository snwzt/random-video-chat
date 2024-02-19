package models

type User struct {
	Username string `json:"username" redis:"username"`
	IPAddr   string `json:"ip_addr" redis:"ip_addr"`
}
