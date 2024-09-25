package httpserver

import (
	"encoding/json"
	"io"
	"log"
	"multiplayer_game/dao"
	"multiplayer_game/service/common/gamedata"
	"net/http"
	"path/filepath"
)

func Initialize() {

	staticFiles := http.FileServer(http.Dir("static"))
	roughJsFiles := http.FileServer(http.Dir("static/roughjsHelper"))

	http.HandleFunc("/", handleCreate)
	http.HandleFunc("/welcome", handleWelcome)
	http.HandleFunc("/output.css", handleTailwindCss)

	http.Handle("/static/", http.StripPrefix("/static/", staticFiles))
	http.Handle("/static/roughjsHelper/", http.StripPrefix("/static/roughjsHelper", roughJsFiles))

	http.HandleFunc("/avatar", handleAvatar)
	http.HandleFunc("/player", handlePlayer)

	http.HandleFunc("/game", handleGame)

	log.Fatal(http.ListenAndServe(":8081", nil))
}

func handleCreate(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("static/game.html"))
}

func handleWelcome(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("static/welcome.html"))
}

func handleTailwindCss(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("static/output.css"))
}

func handleAvatar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", 400)
		return
	}

	if r.Method == http.MethodPost {
		//save base64 data in redis
		base64Data, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error while reading base64 data for avatar %+v", err)
			http.Error(w, "Avatar query failed", 500)
		}
		imageID, err := dao.AddAvatarImage(string(base64Data), gamedata.RedisClient)
		if err != nil {
			log.Printf("Error while creating avatar %+v", err)
			http.Error(w, "Download Failed", 500)
		}
		w.Write([]byte(imageID))
		return
	}

	if r.Method == http.MethodGet {
		fileKey := r.URL.Query().Get("key")
		if fileKey == "" {
			http.Error(w, "Missing query parameter 'key'", http.StatusBadRequest)
			return
		}
		base64Data, err := dao.GetAvatarImage(fileKey, gamedata.RedisClient)
		if err != nil {
			log.Printf("Error while getting key, %+v", err)
			http.Error(w, "Retrieve Error", 500)
			return
		}
		w.Write([]byte(base64Data))
		return
	}
}

func handlePlayer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", 400)
		return
	}

	conId := r.URL.Query().Get("key")
	if conId == "" {
		http.Error(w, "missing key for connectionId", 400)
		return
	}

	player, err := dao.GetPlayerByConnectionId(conId, gamedata.RedisClient)
	if err != nil {
		log.Printf("error while getting player %+v", err)
		http.Error(w, "error while getting player", 500)
		return
	}
	playerBytes, err := json.Marshal(player)
	if err != nil {
		log.Printf("error while json marshaling %+v", err)
		http.Error(w, "error while processing player", 500)
		return
	}
	w.Write(playerBytes)
}

func handleGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", 400)
		return
	}
	gameId := r.URL.Query().Get("key")
	if gameId == "" {
		http.Error(w, "missing param key", 400)
		return
	}
	game, err := dao.GetGameByGameId(gameId, gamedata.RedisClient)
	if err != nil {
		log.Printf("Error while getting game from gameId %+v", err)
		http.Error(w, "Error while getting game from gameId", 500)
		return
	}
	gBytes, err := json.Marshal(game)
	if err != nil {
		log.Printf("Error while marshaling for game %+v", err)
		http.Error(w, "Error while pprocessing game", 500)
		return
	}
	w.Write(gBytes)
}
