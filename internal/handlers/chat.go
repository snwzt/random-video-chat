package handlers

import (
	"context"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type ChatHTTPHandle struct {
	Redis *redis.Client
}

func (h *ChatHTTPHandle) CheckHealth(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}

func (h *ChatHTTPHandle) Chat(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	ws, err := Upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	outMsgChannel := id + ":outgoing"
	incMsgChannel := id + ":incoming"

	var wg sync.WaitGroup
	ch := make(chan bool)

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			msgType, message, err := ws.ReadMessage()
			if err != nil {
				log.Println(err)
			}

			log.Println(msgType)

			if msgType == -1 { // ws closing
				ch <- true

				err = h.Redis.Publish(context.Background(), outMsgChannel, "unavaliable").Err()
				if err != nil {
					log.Println(err)
				}

				if err := h.Redis.LPush(context.Background(), "deleteuser_queue", id).Err(); err != nil {
					log.Println(err)
				}

				break
			}

			// Publish the received message to a Redis Pub/Sub channel
			err = h.Redis.Publish(context.Background(), outMsgChannel, string(message)).Err()
			if err != nil {
				log.Println(err)
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
					log.Println(err)
				}

				// Send the message to the WebSocket
				err = ws.WriteMessage(websocket.TextMessage, []byte(msg.Payload))
				if err != nil {
					log.Println(err)
				}
			}
		}
	}()

	wg.Wait()

	return nil
}
