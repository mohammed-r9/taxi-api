package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	app := application{
		addr: fmt.Sprintf(":%s", port),
	}

	app.mount()
	app.run()
}
