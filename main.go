package main

import (
	"forum/backend"
	"log"
	"net/http"
)

func main() {
	backend.Init()

	log.Println("Forum running → http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", backend.NewRouter()))
}
