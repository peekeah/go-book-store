package main

import "github.com/peekeah/book-store/app"

func main() {
	server := app.InitilizeServer(3000)
	server.Run()
}
