package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for game clients
	},
}

// FrameData represents a game frame update
type FrameData struct {
	PlayerID  string                 `json:"player_id"`
	FrameNum  int64                  `json:"frame_num"`
	Timestamp int64                  `json:"timestamp"`
	Actions   map[string]interface{} `json:"actions"`
}

// Player represents a connected player
type Player struct {
	ID   string
	Conn *websocket.Conn
	Send chan []byte
}

// GameRoom manages frame synchronization for players
type GameRoom struct {
	Players    map[string]*Player
	Broadcast  chan []byte
	Register   chan *Player
	Unregister chan *Player
	mu         sync.RWMutex
	FrameNum   int64
}

func NewGameRoom() *GameRoom {
	return &GameRoom{
		Players:    make(map[string]*Player),
		Broadcast:  make(chan []byte, 256),
		Register:   make(chan *Player),
		Unregister: make(chan *Player),
		FrameNum:   0,
	}
}

func (room *GameRoom) Run() {
	ticker := time.NewTicker(16 * time.Millisecond) // ~60 FPS
	defer ticker.Stop()

	for {
		select {
		case player := <-room.Register:
			room.mu.Lock()
			room.Players[player.ID] = player
			room.mu.Unlock()
			log.Printf("Player %s connected. Total players: %d", player.ID, len(room.Players))

		case player := <-room.Unregister:
			room.mu.Lock()
			if _, ok := room.Players[player.ID]; ok {
				delete(room.Players, player.ID)
				close(player.Send)
				log.Printf("Player %s disconnected. Total players: %d", player.ID, len(room.Players))
			}
			room.mu.Unlock()

		case message := <-room.Broadcast:
			room.mu.RLock()
			for _, player := range room.Players {
				select {
				case player.Send <- message:
				default:
					close(player.Send)
					delete(room.Players, player.ID)
				}
			}
			room.mu.RUnlock()

		case <-ticker.C:
			room.mu.Lock()
			room.FrameNum++
			room.mu.Unlock()
		}
	}
}

func (player *Player) ReadPump(room *GameRoom) {
	defer func() {
		room.Unregister <- player
		player.Conn.Close()
	}()

	player.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	player.Conn.SetPongHandler(func(string) error {
		player.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := player.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		// Parse and re-broadcast frame data
		var frameData FrameData
		if err := json.Unmarshal(message, &frameData); err == nil {
			frameData.Timestamp = time.Now().UnixNano()
			if data, err := json.Marshal(frameData); err == nil {
				room.Broadcast <- data
			}
		}
	}
}

func (player *Player) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		player.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-player.Send:
			player.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				player.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := player.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to current websocket message
			n := len(player.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-player.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			player.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := player.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func serveWs(room *GameRoom, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	playerID := r.URL.Query().Get("player_id")
	if playerID == "" {
		playerID = "player_" + time.Now().Format("20060102150405")
	}

	player := &Player{
		ID:   playerID,
		Conn: conn,
		Send: make(chan []byte, 256),
	}

	room.Register <- player

	go player.WritePump()
	go player.ReadPump(room)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	room := NewGameRoom()
	go room.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(room, w, r)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "ok",
			"players": len(room.Players),
		})
	})

	log.Printf("Game frame sync server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
