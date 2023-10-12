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
	"fmt"
	"github.com/gocolly/colly/v2"
	"github.com/gorilla/websocket"
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
		// Verifica se l'origine della richiesta è consentita
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
				fmt.Println("Errore nell'invio del messaggio:", err)
				client.Close()
				delete(room.Clients, client)
			}
		}

		time.Sleep(1 * time.Second)
	}

}

func setTimer(room *Room) {
	// Invia il valore del timer a tutti i client connessi
	for client := range room.Clients {
		err := client.WriteJSON(WsJson{room.Timer, room.Word})
		if err != nil {
			fmt.Println("Errore nell'invio del messaggio:", err)
			client.Close()
			delete(room.Clients, client)
		}
	}
}

func handleConnections(room *Room) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer conn.Close()

		room.Clients[conn] = true

		// Inizializza il nuovo client con il valore corrente del timer
		err = conn.WriteJSON(WsJson{room.Timer, room.Word})
		if err != nil {
			fmt.Println("Errore nell'invio del messaggio connection:", err)
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

			// Gestisci l'input del client (es. avvio, metti in pausa, ripristina il timer)
			room.Mutex.Lock()
			if msg == 1 {
				// Avvia il timer
				if !room.Active {
					room.Active = true
					room.Word = getWords(room.EasyMode)
					room.Cond.Broadcast()
				}
			} else if msg == 2 {
				// Metti in pausa il timer
				if room.Active {
					room.Active = false
					room.Cond.Broadcast()
				}
			} else if msg == 3 {
				// Ripristina il timer
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
	// Crea un nuovo colly collector
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

func main() {
	room := NewRoom(MAX_TIME, "69420", true)
	http.HandleFunc("/ws", handleConnections(room))
	go handleTimer(room)
	http.ListenAndServe(":4200", nil)
}
