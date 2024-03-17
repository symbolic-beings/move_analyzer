package main

import (
	"fmt"
	"log"
	"move_analyzer/server"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/ws", server.NewAnalysisHandler().Start)

	log.Println("starting the server")
	err := http.ListenAndServe("localhost:3333", nil)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
