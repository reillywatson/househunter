package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}
	http.Handle("/houses", http.HandlerFunc(getHouses))
	http.ListenAndServe(":"+port, nil)

	router.Run(":" + port)
}
