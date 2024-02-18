package handlers

import (
	"context"
	"encoding/json"
	"log"
	"snwzt/random-video-chat/data/models"

	"github.com/redis/go-redis/v9"
)

type ForwarderHandle struct {
	Redis           *redis.Client
	CancelForwarder chan string
}

func (h *ForwarderHandle) CreateMatch() { // queuing
	for {
		matchRequestJSON, err := h.Redis.BRPop(context.Background(), 0, "creatematch_queue").Result()
		if err != nil {
			log.Println(err)
			return
		}

		var matchRequest models.MatchRequest

		log.Println(matchRequestJSON[1])

		if err := json.Unmarshal([]byte(matchRequestJSON[1]), &matchRequest); err != nil {
			log.Println(err)
		}

		log.Println(matchRequest.MatchData.UserID1, matchRequest.MatchData.UserID2)

		User1IncMsgChannel := matchRequest.MatchData.UserID1 + ":incoming"
		User1OutMsgChannel := matchRequest.MatchData.UserID1 + ":outgoing"
		User2IncMsgChannel := matchRequest.MatchData.UserID2 + ":incoming"
		User2OutMsgChannel := matchRequest.MatchData.UserID2 + ":outgoing"

		go func(MatchID string, User1Inc string, User1Out string, User2Inc string, User2Out string) {
			user1Inc := h.Redis.Subscribe(context.Background(), User1Inc)
			user1Out := h.Redis.Subscribe(context.Background(), User1Out)
			user2Inc := h.Redis.Subscribe(context.Background(), User2Inc)
			user2Out := h.Redis.Subscribe(context.Background(), User2Out)

			defer func() {
				user1Inc.Close()
				user1Out.Close()
				user2Inc.Close()
				user2Out.Close()
			}()

			// user1out -> user2inc
			// user2out -> user1inc
			for {
				select {
				case <-h.CancelForwarder:
					log.Println("forwarder" + MatchID + "has been removed")
					return
				case msg, ok := <-user1Out.Channel():
					if !ok {
						log.Println("chat channel " + User1Out + " closed unexpectedly")
						return
					}
					log.Println(msg.Payload)

					h.Redis.Publish(context.Background(), User2Inc, msg.Payload)
				case msg, ok := <-user2Out.Channel():
					if !ok {
						log.Println("chat channel " + User2Out + " closed unexpectedly")
						return
					}
					log.Println(msg.Payload)

					h.Redis.Publish(context.Background(), User1Inc, msg.Payload)
				}
			}

		}(matchRequest.ID, User1IncMsgChannel, User1OutMsgChannel, User2IncMsgChannel, User2OutMsgChannel)

		log.Println("Created " + matchRequest.ID + " forwarder")
	}
}

func (h *ForwarderHandle) DeleteMatch() { // pubsub
	ch := h.Redis.Subscribe(context.Background(), "deletematch_channel")
	defer ch.Close()

	for msg := range ch.Channel() {
		h.CancelForwarder <- msg.Payload
	}
}
