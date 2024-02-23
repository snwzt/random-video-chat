package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"snwzt/rvc/data/models"
	"snwzt/rvc/internal/helpers"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type UserServerHandle struct {
	Redis  *redis.Client
	Logger *zerolog.Logger
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

	if err := h.Redis.HSet(context.Background(), fmt.Sprintf("userentry:%s", userid),
		"username", username, "ipaddr", c.RealIP(), "matchid", "").Err(); err != nil {
		h.Logger.Err(err).Msg("unable to add user entry")
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	if err := h.Redis.SAdd(context.Background(), "unpairedpool", userid).Err(); err != nil {
		h.Logger.Err(err).Msg("unable to add user to unpaired pool")
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	c.SetCookie(&http.Cookie{
		Name:  "rvcuserid",
		Value: userid,
	})

	WsAddr := "wss://" + c.Request().Host + "/chat/" + userid
	TurnUrl := os.Getenv("TURN_URL")
	TurnUser := os.Getenv("TURN_USERNAME")
	TurnCred := os.Getenv("TURN_CRED")

	return c.Render(http.StatusOK, "chat", map[string]string{
		"WsAddr":   WsAddr,
		"TurnUrl":  TurnUrl,
		"TurnUser": TurnUser,
		"TurnCred": TurnCred,
	})
}

func (h *UserServerHandle) MatchUser(c echo.Context) error {
	userIDCookie, err := c.Cookie("rvcuserid")
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	userid1 := userIDCookie.Value

	if err := helpers.RemoveOldMatch(h.Redis, userid1); err != nil {
		h.Logger.Err(err).Msg("unable to remove old match of the user")
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	for attempt := 1; attempt <= 5; attempt++ {
		userid2, err := h.Redis.SRandMemberN(context.Background(), "unpairedpool", 1).Result()
		if err != nil {
			h.Logger.Err(err).Msg("unable to match user")
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

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
			h.Logger.Err(err).Msg("unable to marshal match")
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		if err := h.Redis.LPush(context.Background(), "matchqueue", matchJSON).Err(); err != nil {
			h.Logger.Err(err).Msg("unable to push match in matchqueue")
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		break
	}

	return c.NoContent(http.StatusOK)
}
