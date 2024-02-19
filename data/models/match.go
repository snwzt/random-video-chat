package models

type Match struct {
	ID      string `json:"match_id"`
	UserID1 string `json:"user_id1"`
	UserID2 string `json:"user_id2"`
}
