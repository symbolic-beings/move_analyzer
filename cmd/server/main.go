package main

import (
	"fmt"
	"log"
	"move_analyzer/analyzer"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

const (
	depth = 18
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// TODO: don't do this
	CheckOrigin: func(r *http.Request) bool { return true },
}

func reader(conn *websocket.Conn) {
	done := make(chan interface{})
	s := analyzer.NewStockfish()
	positions := make(chan string)
	out, err := s.StartAnalysis(done, positions, depth)
	if err != nil {
		log.Println(err)
		return
	}
	go writeAnalysisToOutput(out, conn, 1)

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			close(done)
			return
		}

		log.Println("message received: ", string(p))
		positions <- string(p)

	}
}

func writeAnalysisToOutput(out <-chan string, conn *websocket.Conn, messageType int) {
	for analysis := range out {
		fmt.Println("writing to conn")
		if err := conn.WriteMessage(messageType, []byte(analysis)); err != nil {
			log.Println(err)
			return
		}

	}

	fmt.Println("closing the analysis")
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	log.Println("Client Connected")
	err = ws.WriteMessage(1, []byte("you have connected"))
	if err != nil {
		log.Println(err)
	}

	reader(ws)
}

func main() {
	http.HandleFunc("/ws", wsEndpoint)
	err := http.ListenAndServe("localhost:3333", nil)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
