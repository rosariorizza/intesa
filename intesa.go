/*package main

import (
	"net/http"
































	"github.com/gocolly/colly/v2"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {

	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"}, // Permetti tutte le origini
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodOptions},
	}))

	// Routes
	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, getWords(true))


			//var res string
			//for _, s := range getWords(true) {
			//	//fmt.Println("%s")
			//	res += s + "\n"
			//}
			//return c.String(http.StatusOK, res)
	})

	// Start server
	e.Logger.Fatal(e.Start(":8000"))

}

func getWords(easy bool) [10]string {
	// Crea un nuovo colly collector
	c := colly.NewCollector()

	var words [10]string
	var i = -1

	c.OnHTML("div", func(e *colly.HTMLElement) {

		value := e.Attr("div")

		if i < 0 {
			i++
		} else if i < 10 {
			words[i] = e.Text[1:]
			//fmt.Printf("%s\n", e.Text[1:])
			i++
		}

		c.Visit(e.Request.AbsoluteURL(value))
	})

	if easy {
		c.Visit("https://www.parolecasuali.it/?fs=10&fs2=1&Submit=Nuova+parola")
	} else {
		c.Visit("https://www.parolecasuali.it/?fs=10&fs2=0&Submit=Nuova+parola")
	}

	return words
}*/

package main

import (
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly/v2"
	"github.com/gorilla/websocket"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

var MAX_TIME = 60

type WsJson struct {
	Timer int    `json:"timerValue"`
	Word  string `json:"wordValue"`
}

type Room struct {
	RoomCode string
	Timer    int
	Clients  map[*websocket.Conn]bool
	Mutex    *sync.Mutex
	Cond     *sync.Cond
	Active   bool
	EasyMode bool
	Word     string
}

func NewRoom(maxTime int, roomCode string, easyMode bool) *Room {

	mu := sync.Mutex{}
	cond := sync.NewCond(&mu)

	return &Room{
		RoomCode: roomCode,
		Timer:    maxTime,
		Clients:  make(map[*websocket.Conn]bool),
		Mutex:    &mu,
		Cond:     cond,
		Active:   false,
		EasyMode: easyMode,
		Word:     "",
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		allowedOrigins := []string{"null", "http://localhost:3000"}
		origin := r.Header.Get("Origin")
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				return true
			}
		}
		return false
	},
}

func handleTimer(room *Room) {
	for {

		room.Mutex.Lock()
		for !room.Active || room.Timer == 0 {
			room.Cond.Wait()
		}

		room.Mutex.Unlock()

		room.Timer--

		// Invia il valore del timer a tutti i client connessi
		for client := range room.Clients {
			err := client.WriteJSON(WsJson{room.Timer, room.Word})
			if err != nil {
				fmt.Println("Error sending the message: ", err)
				client.Close()
				delete(room.Clients, client)
			}
		}

		time.Sleep(1 * time.Second)
	}

}

func setTimer(room *Room) {
	for client := range room.Clients {
		err := client.WriteJSON(WsJson{room.Timer, room.Word})
		if err != nil {
			fmt.Println("Error sending the message: ", err)
			client.Close()
			delete(room.Clients, client)
		}
	}
}

func handleConnections(rooms map[string]*Room) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL.Query())
		room, ok := rooms[r.URL.Query().Get("room")]
		if !ok {
			fmt.Println("Room not found")
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer conn.Close()

		room.Clients[conn] = true

		err = conn.WriteJSON(WsJson{room.Timer, room.Word})
		if err != nil {
			fmt.Println("Error sending the message: ", err)
			return
		}

		for {
			var msg int
			err := conn.ReadJSON(&msg)
			if err != nil {
				fmt.Println(err)
				delete(room.Clients, conn)
				return
			}

			room.Mutex.Lock()
			if msg == 1 {
				// Start timer
				if !room.Active {
					room.Active = true
					room.Word = getWords(room.EasyMode)
					room.Cond.Broadcast()
				}
			} else if msg == 2 {
				// Stop timer
				if room.Active {
					room.Active = false
					room.Cond.Broadcast()
				}
			} else if msg == 3 {
				// Reset timer
				room.Timer = MAX_TIME
				room.Word = ""
				setTimer(room)
				room.Active = false
				room.Cond.Broadcast()
			}
			room.Mutex.Unlock()

		}
	}
}

func getWords(easy bool) string {
	c := colly.NewCollector()

	var word string

	c.OnHTML("div", func(e *colly.HTMLElement) {

		value := e.Attr("div")

		word = e.Text[1:]

		c.Visit(e.Request.AbsoluteURL(value))
	})

	if easy {
		c.Visit("https://www.parolecasuali.it/?fs=1&fs2=1&Submit=Nuova+parola")
	} else {
		c.Visit("https://www.parolecasuali.it/?fs=1&fs2=0&Submit=Nuova+parola")
	}

	return word
}

func StartRoom(room *Room) {
}

const letterBytes = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func main() {

	rooms := make(map[string]*Room)

	http.HandleFunc("/newRoom", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		var roomCode string
		for {
			roomCode = RandStringBytes(6)
			if _, ok := rooms[roomCode]; !ok {
				break
			}
		}

		room := NewRoom(MAX_TIME, roomCode, true)

		go handleTimer(room)

		rooms[roomCode] = room

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(roomCode)
	})

	http.HandleFunc("/play", handleConnections(rooms))

	http.ListenAndServe(":4200", nil)
}
