package dao

import (
	"encoding/json"
	"fmt"
	"multiplayer_game/dto"
	"multiplayer_game/util"
	"time"

	"github.com/go-redis/redis"
)

func AddAvatarImage(base64Data string, redisClient *redis.Client) (string, error) {
	imageId := util.GenerateGuid()
	err := redisClient.Set(imageId, base64Data, time.Hour).Err()
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

func AddPlayer(player *dto.Player, key string, redisClient *redis.Client) error {
	return redisClient.Set(key, player, time.Hour).Err()
}

func GetEntity(key string, redisClient *redis.Client) (string, error) {
	strPlayer, err := redisClient.Get(key).Result()
	if err != nil {
		return "", err
	}
	return strPlayer, nil
}

// Add player to particular GameId
func AddPlayerByGameId(player *dto.Player, gameId string, connectionId string, redisClient *redis.Client, join bool) error {

	// return error if Player variable lacks some necessary attributes
	if len(player.Name) == 0 {
		return fmt.Errorf("name of player is invalid")
	}
	//get the game already sotred in redis
	game, err := GetGameByGameId(gameId, redisClient)
	if err == redis.Nil {
		if join {
			return err
		}
		game = &dto.GameState{
			TotalRounds:        5,
			CurrRound:          0,
			Complete:           false,
			Start:              false,
			TotalPlayers:       1,
			ActivePlayers:      []string{connectionId},
			CurrPlayerScoreMap: make(map[string]uint16),
			PlayerScoreMap:     make(map[string]uint16),
			InactivePlayers:    make([]string, 0),
			AlreadyDrawn:       make([]string, 0),
		}
		return redisClient.Set("GAME_"+gameId, game, time.Hour).Err()
	}
	if err != nil {
		return err
	}
	//check if the name or id of player is already taken
	aPlayers := append(game.ActivePlayers, game.InactivePlayers...)
	for _, conId := range aPlayers {
		if connectionId == conId {
			return fmt.Errorf("already added to the game")
		}
		tPlayer, err := GetPlayerByConnectionId(conId, redisClient)
		if err != nil {
			continue
		}
		if tPlayer.Name == player.Name {
			return fmt.Errorf("name %s already taken for the game", player.Name)
		}
	}

	game.ActivePlayers = append(game.ActivePlayers, connectionId)

	return redisClient.Set("GAME_"+gameId, game, time.Hour).Err()
}

// Add player to particular ConnectionId
func AddPlayerByConnectionId(player *dto.Player, connectionId string, redisClient *redis.Client) error {
	return AddPlayer(player, "CON_PLAYER_"+connectionId, redisClient)
}

// Get player By particular ConnectionId
func GetPlayerByConnectionId(connectionId string, redisClient *redis.Client) (*dto.Player, error) {
	var result string
	result, err := GetEntity("CON_PLAYER_"+connectionId, redisClient)
	if err != nil {
		return nil, err
	}
	var player dto.Player
	err = json.Unmarshal([]byte(result), &player)
	if err != nil {
		return nil, err
	}
	return &player, nil
}

// Get game by particular gameId
func GetGameByGameId(gameId string, redisClient *redis.Client) (*dto.GameState, error) {
	result, err := GetEntity("GAME_"+gameId, redisClient)
	if err != nil {
		return nil, err
	}
	var game dto.GameState
	err = json.Unmarshal([]byte(result), &game)
	if err != nil {
		return nil, err
	}
	return &game, nil
}

func MakeConnectionIdIncactive(gameId string, conId string, redisClient *redis.Client) error {
	game, err := GetGameByGameId(gameId, redisClient)
	if err != nil {
		return err
	}
	activePlayers := game.ActivePlayers
	needToUpdate := false

	var newActive []string = make([]string, 0)

	for _, playerId := range activePlayers {
		if conId == playerId {
			needToUpdate = true
		} else {
			newActive = append(newActive, playerId)
		}
	}
	if !needToUpdate {
		return fmt.Errorf("No connectionId %s in game %s", conId, gameId)
	}

	game.ActivePlayers = newActive
	game.InactivePlayers = append(game.InactivePlayers, conId)

	return redisClient.Set("GAME_"+gameId, game, time.Hour).Err()
}
