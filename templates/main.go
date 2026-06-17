package main

import (
	"html/template"
	"log"
	"net/http"
)

func main() {
	initDB()
	defer db.Close()

	var err error
	tmpl, err = template.New("").Funcs(funcMap).ParseGlob("templates/*.html")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Forum running → http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", newRouter()))
}
