package main

import (
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

const (
	startingPosition = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
)

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: "localhost:3333", Path: "/ws"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial: ", err.Error())
	}

	defer c.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			mtype, msg, err := c.ReadMessage()
			if err != nil {
				log.Println("read: ", err.Error())
			}

			log.Printf("recv: %s, type: %d", msg, mtype)
		}
	}()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			log.Println("interrupt")

			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close: ", err.Error())
				return
			}

			return
		case <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte(startingPosition))
			if err != nil {
				log.Println("write: ", err.Error())
				return
			}
		}
	}
}
