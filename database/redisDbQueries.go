package database

import (
	"multiplayer_game/util"
	"time"

	"github.com/go-redis/redis"
)

func AddAvatarImage(base64Data string, redisClient *redis.Client) (string, error) {
	imageId := util.GenerateGuid()
	err := redisClient.Set(imageId, base64Data, time.Minute*5).Err()
	if err != nil {
		return "", err
	}
	return imageId, nil
}

func GetAvatarImage(imageID string, redisClient *redis.Client) (string, error) {
	imageIdVAlue, err := redisClient.Get(imageID).Result()
	if err != nil {
		return "", err
	}
	return imageIdVAlue, nil
}
