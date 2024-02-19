package handlers

import (
	"context"
	"encoding/json"
	"log"
	"snwzt/rvc/data/models"
	"snwzt/rvc/internal/helpers"

	"github.com/redis/go-redis/v9"
)

type ForwarderOperationsHandle struct {
	Redis           *redis.Client
	CancelForwarder chan string
}

func (h *ForwarderOperationsHandle) CreateMatch() { // queue
	for {
		matchJSON, err := h.Redis.BRPop(context.Background(), 0, "creatematch").Result()
		if err != nil {
			log.Println(err)
			return
		}

		var match models.Match

		log.Println(matchJSON[1])

		if err := json.Unmarshal([]byte(matchJSON[1]), &match); err != nil {
			log.Println(err)
			continue
		}

		log.Println(match.UserID1, match.UserID2)

		go helpers.Forwarder(h.CancelForwarder, match, h.Redis)

		log.Println("Created " + match.ID + " forwarder")
	}
}

func (h *ForwarderOperationsHandle) DeleteMatch() { // pubsub
	ch := h.Redis.Subscribe(context.Background(), "deletematch")
	defer ch.Close()

	for msg := range ch.Channel() {
		log.Println("recieved cancellation request for ", msg.Payload)
		h.CancelForwarder <- msg.Payload
	}
}
