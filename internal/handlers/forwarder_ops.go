package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"snwzt/rvc/data/models"
	"snwzt/rvc/internal/helpers"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type ForwarderOperationsHandle struct {
	Redis           *redis.Client
	CancelForwarder chan string
	Logger          *zerolog.Logger
}

func (h *ForwarderOperationsHandle) CreateMatch() { // queue
	for {
		matchJSON, err := h.Redis.BRPop(context.Background(), 0, "creatematch").Result()
		if err != nil {
			h.Logger.Err(err).Msg("unable to pop match from creatematch queue")
			continue
		}

		var match models.Match

		if err := json.Unmarshal([]byte(matchJSON[1]), &match); err != nil {
			h.Logger.Err(err).Msg("unable to unmarshal match")
			continue
		}

		go helpers.Forwarder(h.CancelForwarder, match, h.Redis, h.Logger)

		h.Logger.Info().Msg(fmt.Sprintf("created %s for %s %s", match.ID, match.UserID1, match.UserID2))
	}
}

func (h *ForwarderOperationsHandle) DeleteMatch() { // pubsub
	ch := h.Redis.Subscribe(context.Background(), "deletematch")
	defer ch.Close()

	for msg := range ch.Channel() {
		h.Logger.Info().Msg("recieved cancellation request for " + msg.Payload)
		h.CancelForwarder <- msg.Payload
	}
}
