package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/peekeah/book-store/app"
)

func main() {
	// Load env
	if err := godotenv.Load(); err != nil {
		log.Fatal("error while loading .env")
	}

	server := app.NewSever()
	server.MigragateDB()
	server.Run()
}
