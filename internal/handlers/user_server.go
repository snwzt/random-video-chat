package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"snwzt/rvc/data/models"
	"snwzt/rvc/internal/helpers"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

type UserServerHandle struct {
	Redis *redis.Client
}

func (h *UserServerHandle) CheckHealth(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}

func (h *UserServerHandle) Home(c echo.Context) error {
	return c.Render(http.StatusOK, "index.html", nil)
}

func (h *UserServerHandle) RegisterUser(c echo.Context) error {
	username := c.FormValue("username")
	userid := "user:" + username + ":" + uuid.New().String()

	if err := h.Redis.SAdd(context.Background(), "unpairedpool", userid).Err(); err != nil {
		log.Println(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	if err := h.Redis.HSet(context.Background(), fmt.Sprintf("userentry:%s", userid),
		"username", username, "ipaddr", c.RealIP(), "matchid", "").Err(); err != nil {
		log.Println(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	c.SetCookie(&http.Cookie{
		Name:  "rvcuserid",
		Value: userid,
	})

	wsaddr := "ws://" + "localhost:5001" + "/chat/" + userid

	return c.Render(http.StatusOK, "chat", map[string]string{
		"wsAddr": wsaddr,
	})
}

func (h *UserServerHandle) MatchUser(c echo.Context) error {
	userIDCookie, err := c.Cookie("rvcuserid")
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	userid1 := userIDCookie.Value

	if err := helpers.RemoveOldMatch(h.Redis, userid1); err != nil {
		log.Println(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	userid2, err := h.Redis.SRandMemberN(context.Background(), "unpairedpool", 1).Result()
	if err != nil {
		log.Println(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	for attempt := 1; attempt <= 5; attempt++ {
		if userid1 == userid2[0] {
			time.Sleep(time.Duration(attempt) * 2 * time.Second)
			continue
		}

		matchJSON, err := json.Marshal(&models.Match{
			ID:      fmt.Sprintf("match:%s", uuid.New().String()),
			UserID1: userid1,
			UserID2: userid2[0],
		})
		if err != nil {
			log.Println(err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		if err := h.Redis.LPush(context.Background(), "matchqueue", matchJSON).Err(); err != nil {
			log.Println(err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		break
	}

	return c.NoContent(http.StatusOK)
}
