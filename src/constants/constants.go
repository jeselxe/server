package constants

import (
	"log"
	"os"
)

// URI uri from environment
var URI = getURI()

// AuthDatabase database from environment
var AuthDatabase = getDatabase()

// Port port from environment
var Port = GetPort()

var TcpPort = GetTcpPort()

func getURI() string {
	var uri = os.Getenv("MONGOLAB_URI")
	if uri == "" {
		uri = "localhost:27017"
		log.Println("INFO: URI taken from env")
	}
	return uri
}

func getDatabase() string {
	var db = os.Getenv("MONGOLAB_DB")
	if db == "" {
		db = "securechat"
		log.Println("INFO: Database taken from env")
	}
	return db
}

// GetPort Get the Port from the environment so we can run on Heroku
func GetPort() string {
	var port = os.Getenv("PORT")
	// Set a default port if there is nothing in the environment
	if port == "" {
		port = "8080"
		log.Println("INFO: No PORT environment variable detected, defaulting to " + port)
	}
	return ":" + port
}

// GetPort Get the Port from the environment so we can run on Heroku
func GetTcpPort() string {
	var port = os.Getenv("TCP_PORT")
	// Set a default port if there is nothing in the environment
	if port == "" {
		port = "1337"
		log.Println("INFO: No TCP_PORT environment variable detected, defaulting to " + port)
	}
	return ":" + port
}
