package server

import (
	"fmt"
	"log"
	"move_analyzer/analyzer"
	"net/http"

	"github.com/gorilla/websocket"
)

type Socket interface {
	Start(w http.ResponseWriter, r *http.Request)
}

const (
	depth = 18
)

type analysisHandler struct {
	upgrader websocket.Upgrader
}

func NewAnalysisHandler() Socket {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		// TODO: don't do this
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	return &analysisHandler{upgrader}
}

func (a *analysisHandler) Start(w http.ResponseWriter, r *http.Request) {
	ws, err := a.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	log.Println("Handling connection from client")
	err = ws.WriteMessage(1, []byte("you have connected"))
	if err != nil {
		log.Println(err)
	}

	a.monitor(ws)
}

func (a *analysisHandler) monitor(conn *websocket.Conn) {
	incomingPositions := make(chan string)
	done := a.setupMonitor(conn, incomingPositions)
	defer close(incomingPositions)
	defer close(done)

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		log.Println("message received: ", string(p))
		incomingPositions <- string(p)
	}
}

func (a *analysisHandler) setupMonitor(conn *websocket.Conn, incomingPositions <-chan string) chan interface{} {
	done := make(chan interface{})
	s := analyzer.NewStockfish()

	out, err := s.StartAnalysis(done, incomingPositions, depth)
	if err != nil {
		log.Println(err)
		return done
	}

	go writeAnalysisToOutput(out, conn, 1)

	return done
}

func writeAnalysisToOutput(out <-chan string, conn *websocket.Conn, messageType int) {
	for analysis := range out {
		if err := conn.WriteMessage(messageType, []byte(analysis)); err != nil {
			log.Println(err)
			return
		}
	}

	fmt.Println("closing the analysis")
}
