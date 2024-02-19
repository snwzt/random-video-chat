package helpers

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

func RemoveOldMatch(rclient *redis.Client, userid string) error {
	matchid, err := rclient.HGet(context.Background(), fmt.Sprintf("userentry:%s", userid), "matchid").Result()
	if err != nil {
		return err
	}

	if matchid != "" {
		users, err := rclient.HGetAll(context.Background(), fmt.Sprintf("match:%s", matchid)).Result()
		if err != nil && err != redis.Nil {
			return err
		}

		log.Println(users)

		if err != redis.Nil {
			for _, user := range users {
				// put to unpaired pool
				if err := rclient.SAdd(context.Background(), "unpairedpool", user).Err(); err != nil {
					return err
				}
			}

			// delete match entry
			if err := rclient.Del(context.Background(), fmt.Sprintf("match:%s", matchid)).Err(); err != nil {
				return err
			}

			// delete forwarder
			if err := rclient.Publish(context.Background(), "deletematch", matchid).Err(); err != nil {
				return err
			}
		}
	}

	return nil
}
