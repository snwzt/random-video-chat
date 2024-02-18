package models

import "time"

type Match struct {
	UserID1   string    `json:"user_id1"`
	UserID2   string    `json:"user_id2"`
	Timestamp time.Time `json:"timestamp"`
}

type MatchRequest struct {
	ID        string `json:"match_id"`
	MatchData *Match `json:"match_data"`
}
