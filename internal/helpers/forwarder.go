package helpers

import (
	"context"
	"fmt"
	"log"
	"snwzt/rvc/data/models"

	"github.com/redis/go-redis/v9"
)

func Forwarder(cancelChan chan string, match models.Match, rclient *redis.Client) {
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
		log.Println(err)
	}

	username2, err := rclient.HGet(context.Background(), fmt.Sprintf("userentry:%s", match.UserID2), "username").Result()
	if err != nil {
		log.Println(err)
	}

	if err := rclient.Publish(context.Background(), User1Inc, username2).Err(); err != nil {
		log.Println(err)
	}

	if err := rclient.Publish(context.Background(), User2Inc, username1).Err(); err != nil {
		log.Println(err)
	}

	// user1out -> user2inc
	// user2out -> user1inc
	for {
		select {
		case <-cancelChan:
			log.Println("forwarder" + match.ID + "has been removed")
			return
		case msg, ok := <-user1Out.Channel():
			if !ok {
				log.Println("chat channel " + User1Out + " closed unexpectedly")
				return
			}
			log.Println(msg.Payload)

			if err := rclient.Publish(context.Background(), User2Inc, msg.Payload).Err(); err != nil {
				log.Println(err)
			}
		case msg, ok := <-user2Out.Channel():
			if !ok {
				log.Println("chat channel " + User2Out + " closed unexpectedly")
				return
			}
			log.Println(msg.Payload)

			if err := rclient.Publish(context.Background(), User1Inc, msg.Payload).Err(); err != nil {
				log.Println(err)
			}
		}
	}
}
