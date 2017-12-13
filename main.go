package main

import (
	"log"
	"net/http"
	"os"
)

var mapsApiKey string

func main() {
	port := os.Getenv("PORT")
	mapsApiKey = os.Getenv("GOOGLE_MAPS_KEY")

	if port == "" {
		log.Fatal("$PORT must be set")
	}
	http.Handle("/houses", http.HandlerFunc(getHouses))
	http.ListenAndServe(":"+port, nil)
}
