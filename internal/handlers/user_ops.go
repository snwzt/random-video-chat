package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"snwzt/rvc/data/models"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type UserOperationsHandle struct {
	Redis  *redis.Client
	Logger *zerolog.Logger
}

func (h *UserOperationsHandle) UserMatcher() {
	for {
		matchJSON, err := h.Redis.BRPop(context.Background(), 0, "matchqueue").Result()
		if err != nil {
			h.Logger.Err(err).Msg("unable to pop match from matchqueue")
			return
		}

		var match models.Match

		if err := json.Unmarshal([]byte(matchJSON[1]), &match); err != nil {
			h.Logger.Err(err).Msg("unable to unmarshal match")
			continue
		}

		// check whether user1 and user2 exist in unpaired_pool, if yes -> continue, no -> skip the request
		if !h.Redis.SIsMember(context.Background(), "unpairedpool", match.UserID1).Val() ||
			!h.Redis.SIsMember(context.Background(), "unpairedpool", match.UserID2).Val() {
			continue
		}

		if err := h.Redis.SRem(context.Background(), "unpairedpool", match.UserID1).Err(); err != nil {
			h.Logger.Err(err).Msg("unable to remove user from unpairedpool")
			continue
		}
		if err := h.Redis.SRem(context.Background(), "unpairedpool", match.UserID2).Err(); err != nil {
			h.Logger.Err(err).Msg("unable to remove user from unpairedpool")
			continue
		}

		if err := h.Redis.HSet(context.Background(), fmt.Sprintf("match:%s", match.ID),
			"user1", match.UserID1, "user2", match.UserID2).Err(); err != nil {
			h.Logger.Err(err).Msg("unable to add matchdata")
			continue
		}

		if err := h.Redis.HSet(context.Background(), fmt.Sprintf("userentry:%s", match.UserID1),
			"matchid", match.ID).Err(); err != nil {
			h.Logger.Err(err).Msg("unable to update userentry")
			continue
		}

		if err := h.Redis.HSet(context.Background(), fmt.Sprintf("userentry:%s", match.UserID2),
			"matchid", match.ID).Err(); err != nil {
			h.Logger.Err(err).Msg("unable to update userentry")
			continue
		}

		if err := h.Redis.LPush(context.Background(), "creatematch", matchJSON[1]).Err(); err != nil {
			h.Logger.Err(err).Msg("unable to push to creatematch")
			continue
		}
	}
}

func (h *UserOperationsHandle) UserRemove() {
	for {
		user, err := h.Redis.BRPop(context.Background(), 0, "removeuser").Result()
		if err != nil {
			h.Logger.Err(err).Msg("unable to pop user from removeuser queue")
			return
		}

		userID := user[1]

		matchid, err := h.Redis.HGet(context.Background(), fmt.Sprintf("userentry:%s", userID), "matchid").Result()
		if err != nil {
			h.Logger.Err(err).Msg("unable to get userentry")
			continue
		}

		if !h.Redis.SIsMember(context.Background(), "unpairedpool", userID).Val() {
			// delete forwarder
			if err := h.Redis.Publish(context.Background(), "deletematch", matchid).Err(); err != nil {
				h.Logger.Err(err).Msg("unable to publish to deletematch")
				continue
			}
		}

		// delete match entry
		if err := h.Redis.Del(context.Background(), fmt.Sprintf("match:%s", matchid)).Err(); err != nil {
			h.Logger.Err(err).Msg("unable to delete matchdata")
			continue
		}

		// remove from unpaired pool
		if err := h.Redis.SRem(context.Background(), "unpairedpool", userID).Err(); err != nil {
			h.Logger.Err(err).Msg("unable to remove user from unpairedpool")
			continue
		}

		if err := h.Redis.Del(context.Background(), fmt.Sprintf("userentry:%s", userID)).Err(); err != nil {
			h.Logger.Err(err).Msg("unable to delete userentry")
			continue
		}
	}
}
