package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"snwzt/rvc/data/models"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

func Forwarder(cancelChan chan string, match models.Match, rclient *redis.Client, log *zerolog.Logger) {
	User1Inc := match.UserID1 + ":incoming"
	User1Out := match.UserID1 + ":outgoing"
	User2Inc := match.UserID2 + ":incoming"
	User2Out := match.UserID2 + ":outgoing"

	user1Inc := rclient.Subscribe(context.Background(), User1Inc)
	user1Out := rclient.Subscribe(context.Background(), User1Out)
	user2Inc := rclient.Subscribe(context.Background(), User2Inc)
	user2Out := rclient.Subscribe(context.Background(), User2Out)

	defer func() {
		user1Inc.Close()
		user1Out.Close()
		user2Inc.Close()
		user2Out.Close()
	}()

	// meet and greet
	username1, err := rclient.HGet(context.Background(), fmt.Sprintf("userentry:%s", match.UserID1), "username").Result()
	if err != nil {
		log.Err(err).Msg("unable to get userentry")
	}

	username2, err := rclient.HGet(context.Background(), fmt.Sprintf("userentry:%s", match.UserID2), "username").Result()
	if err != nil {
		log.Err(err).Msg("unable to get userentry")
	}

	msgJSON1, err := json.Marshal(&models.SendMessage{
		Event: "candidate",
		Data: &models.Candidate{
			SDP:      "",
			Username: username2,
		},
	})
	if err != nil {
		log.Err(err).Msg("unable to marshal message")
	}

	if err := rclient.Publish(context.Background(), User1Inc, msgJSON1).Err(); err != nil {
		log.Err(err).Msg("unable to publish to user1inc")
	}

	msgJSON2, err := json.Marshal(&models.SendMessage{
		Event: "candidate",
		Data: &models.Candidate{
			SDP:      "",
			Username: username1,
		},
	})
	if err != nil {
		log.Err(err).Msg("unable to marshal message")
	}

	if err := rclient.Publish(context.Background(), User2Inc, msgJSON2).Err(); err != nil {
		log.Err(err).Msg("unable to publish to user1inc")
	}

	// user1out -> user2inc
	// user2out -> user1inc
	for {
		select {
		case <-cancelChan:
			log.Info().Msg("forwarder " + match.ID + " has been removed")
			return
		case msg, ok := <-user1Out.Channel():
			if !ok {
				log.Info().Msg("chat channel " + User1Out + " closed unexpectedly")
				return
			}

			msgJSON, err := json.Marshal(&models.SendMessage{
				Event: "message",
				Data: &models.Chat{
					Message: msg.Payload,
				},
			})
			if err != nil {
				log.Err(err).Msg("unable to marshal message")
			}

			if err := rclient.Publish(context.Background(), User2Inc, msgJSON).Err(); err != nil {
				log.Err(err).Msg("unable to publish to user2inc")
			}
		case msg, ok := <-user2Out.Channel():
			if !ok {
				log.Info().Msg("chat channel " + User1Out + " closed unexpectedly")
				return
			}

			msgJSON, err := json.Marshal(&models.SendMessage{
				Event: "message",
				Data: &models.Chat{
					Message: msg.Payload,
				},
			})
			if err != nil {
				log.Err(err).Msg("unable to marshal message")
			}

			if err := rclient.Publish(context.Background(), User1Inc, msgJSON).Err(); err != nil {
				log.Err(err).Msg("unable to publish to user1inc")
			}
		}
	}
}
