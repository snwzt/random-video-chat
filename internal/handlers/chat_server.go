package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"snwzt/rvc/data/models"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type ChatServerHandle struct {
	Redis  *redis.Client
	Logger *zerolog.Logger
}

func (h *ChatServerHandle) CheckHealth(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}

func (h *ChatServerHandle) Chat(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	ws, err := Upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("unable to upgrade to websocket: %s", err.Error()))
	}
	defer ws.Close()

	h.Logger.Info().Msg("established websocket conn " + id)

	outMsgChannel := id + ":outgoing"
	incMsgChannel := id + ":incoming"

	var wg sync.WaitGroup
	ch := make(chan bool, 1)

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			msgType, message, err := ws.ReadMessage()
			if err != nil && msgType != -1 {
				h.Logger.Err(err).Msg("unable to read from websocket")
			}

			if msgType == -1 { // ws closing
				ch <- true

				if err := h.Redis.Publish(context.Background(), outMsgChannel, "unavaliable").Err(); err != nil {
					h.Logger.Err(err).Msg("unable to publish to " + outMsgChannel)
				}

				if err := h.Redis.LPush(context.Background(), "removeuser", id).Err(); err != nil {
					h.Logger.Err(err).Msg("unable to push to removeuser queue")
				}

				h.Logger.Info().Msg("closed websocket conn " + id)

				return
			}

			var recieveMessage models.RecieveMessage

			if err := json.Unmarshal(message, &recieveMessage); err != nil {
				h.Logger.Err(err).Msg("unable to unmarshal recievedMessage")
			}

			if recieveMessage.Event == "sdp" {
				log.Println(recieveMessage.Data)

				// update sdp
				if err := h.Redis.HSet(context.Background(), fmt.Sprintf("userentry:%s", id),
					"sdp", recieveMessage.Data).Err(); err != nil {
					h.Logger.Err(err).Msg("unable to update userentry")
				}
			}

			if recieveMessage.Event == "message" {
				// Publish the received message to a Redis Pub/Sub channel
				if err := h.Redis.Publish(context.Background(), outMsgChannel, recieveMessage.Data).Err(); err != nil {
					h.Logger.Err(err).Msg("unable to publish to " + outMsgChannel)
				}
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		pubsub := h.Redis.Subscribe(context.Background(), incMsgChannel)
		defer pubsub.Close()

		for {
			select {
			case <-ch:
				return
			default:
				msg, err := pubsub.ReceiveMessage(context.Background())
				if err != nil {
					h.Logger.Err(err).Msg("unable to get message from" + incMsgChannel)
				}

				// Send the message to the WebSocket
				if err := ws.WriteMessage(websocket.TextMessage, []byte(msg.Payload)); err != nil {
					h.Logger.Err(err).Msg("unable to write to websocket")
				}
			}
		}
	}()

	wg.Wait()

	return nil
}
