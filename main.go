package main

import (
	"log"
	"os"
	"user/database"
	"user/server"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load("./.user-env"); err != nil {
		if _, envExist := os.LookupEnv("ENV_LOADED"); !envExist {
			log.Fatal("environment file not found")
		}
	}

	database.ConnectDb()

	server.RunServer()
}
