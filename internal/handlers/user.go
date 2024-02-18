package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"snwzt/random-video-chat/data/models"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

type UserHTTPHandle struct {
	Redis *redis.Client
}

func (h *UserHTTPHandle) CheckHealth(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}

func (h *UserHTTPHandle) Home(c echo.Context) error {
	return c.Render(http.StatusOK, "index.html", nil)
}

func (h *UserHTTPHandle) RegisterUser(c echo.Context) error {
	username := c.FormValue("username")
	userid := "user:" + uuid.New().String()

	userData := &models.User{
		Username: username,
		IPAddr:   c.RealIP(),
	}

	userDataJSON, err := json.Marshal(userData)
	if err != nil {
		log.Println(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	h.Redis.SAdd(context.Background(), "unpaired_pool", userid)
	h.Redis.HSet(context.Background(), "user_table", userid, userDataJSON)

	c.SetCookie(&http.Cookie{
		Name:  "rvc-usrid",
		Value: userid,
	})

	wsAddr := "ws://" + "localhost:5001" + "/chat/" + userid

	return c.Render(http.StatusOK, "chat", map[string]string{
		"wsAddr": wsAddr,
	})
}

func (h *UserHTTPHandle) MatchUser(c echo.Context) error {
	userIDCookie, err := c.Cookie("rvc-usrid")
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	userid1 := userIDCookie.Value

	// check for old match
	matchid, err := h.Redis.HGet(context.Background(), "usermatch_map", userid1).Result() // update this on other side
	if err != nil && err != redis.Nil {
		log.Println(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	if err != redis.Nil { // matchid exists, user was matched before
		data, err := h.Redis.HGet(context.Background(), "match_table", matchid).Result()
		if err != nil && err != redis.Nil {
			log.Println(err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		log.Println(data)

		if err != redis.Nil {
			var dataJSON models.Match
			if err := json.Unmarshal([]byte(data), &dataJSON); err != nil {
				log.Println(err)
				return echo.NewHTTPError(http.StatusInternalServerError)
			}

			log.Println(dataJSON)

			// put to unpaired pool
			h.Redis.SAdd(context.Background(), "unpaired_pool", dataJSON.UserID1)
			h.Redis.SAdd(context.Background(), "unpaired_pool", dataJSON.UserID2)

			// delete match entry
			h.Redis.HDel(context.Background(), "match_table", matchid)

			// delete forwarder
			h.Redis.Publish(context.Background(), "deletematch_channel", matchid)
		}
	}

	userid2, err := h.Redis.SRandMemberN(context.Background(), "unpaired_pool", 1).Result()
	if err != nil {
		return err
	}

	for attempt := 1; attempt <= 3; attempt++ {
		if userid1 == userid2[0] {
			time.Sleep(time.Duration(attempt) * 10 * time.Second)
			continue
		}

		matchid := "match:" + uuid.New().String()

		matchRequest := &models.MatchRequest{
			ID: matchid,
			MatchData: &models.Match{
				UserID1:   userid1,
				UserID2:   userid2[0],
				Timestamp: time.Now(),
			},
		}

		matchRequestJSON, err := json.Marshal(matchRequest)
		if err != nil {
			log.Println(err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		if err := h.Redis.LPush(context.Background(), "match_request_queue", matchRequestJSON).Err(); err != nil {
			log.Println(err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		break
	}

	return c.NoContent(http.StatusOK)
}

type UserQueueHandle struct {
	Redis *redis.Client
}

func (h *UserQueueHandle) Matcher() {
	for {
		matchRequestJSON, err := h.Redis.BRPop(context.Background(), 0, "match_request_queue").Result()
		if err != nil {
			log.Println(err)
			return
		}

		var matchRequest models.MatchRequest

		if err := json.Unmarshal([]byte(matchRequestJSON[1]), &matchRequest); err != nil {
			log.Println(err)
		}

		log.Println(matchRequest.MatchData.UserID1, matchRequest.MatchData.UserID2)

		// check whether user1 and user2 exist in unpaired_pool, if yes -> continue, no -> skip the request
		checkUser1 := h.Redis.SIsMember(context.Background(), "unpaired_pool", matchRequest.MatchData.UserID1).Val()
		checkUser2 := h.Redis.SIsMember(context.Background(), "unpaired_pool", matchRequest.MatchData.UserID2).Val()

		if !checkUser1 || !checkUser2 {
			continue
		}

		if err := h.Redis.SRem(context.Background(), "unpaired_pool", matchRequest.MatchData.UserID1).Err(); err != nil {
			log.Println(err)
			continue
		}
		if err := h.Redis.SRem(context.Background(), "unpaired_pool", matchRequest.MatchData.UserID2).Err(); err != nil {
			log.Println(err)
			continue
		}

		matchdataJSON, err := json.Marshal(matchRequest.MatchData)
		if err != nil {
			log.Println(err)
			continue
		}

		h.Redis.HSet(context.Background(), "match_table", matchRequest.ID, matchdataJSON)
		h.Redis.HSet(context.Background(), "usermatch_map", matchRequest.MatchData.UserID1, matchRequest.ID)
		h.Redis.HSet(context.Background(), "usermatch_map", matchRequest.MatchData.UserID2, matchRequest.ID)

		matchCreateJSON, err := json.Marshal(matchRequest)
		if err != nil {
			log.Println(err)
			continue
		}

		if err := h.Redis.LPush(context.Background(), "creatematch_queue", matchCreateJSON).Err(); err != nil {
			log.Println(err)
		}

		log.Println(matchRequestJSON)
	}
}

func (h *UserQueueHandle) UserRemove() {
	for {
		user, err := h.Redis.BRPop(context.Background(), 0, "deleteuser_queue").Result()
		if err != nil {
			log.Println(err)
			return
		}

		userID := user[1]

		if err := h.Redis.SRem(context.Background(), "unpaired_pool", userID).Err(); err != nil {
			log.Println(err)
			continue
		}

		if err := h.Redis.HDel(context.Background(), "user_table", userID).Err(); err != nil {
			log.Println(err)
			continue
		}

		matchid, err := h.Redis.HGet(context.Background(), "usermatch_map", userID).Result()
		if err != nil && err != redis.Nil {
			log.Println(err)
			continue
		}

		if err := h.Redis.HDel(context.Background(), "usermatch_map", userID).Err(); err != nil {
			log.Println(err)
			continue
		}

		// delete forwarder
		h.Redis.Publish(context.Background(), "deletematch_channel", matchid)
	}
}
