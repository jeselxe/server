package src

import (
	"os"
)

var URI = getURI()
var AUTH_DATABASE = getDatabase()
var PORT = GetPort()

func getURI() string {
    var uri = os.Getenv("MONGOLAB_URI")
	if uri == "" {
        uri = "localhost:27017"
        log.Println("INFO: dev_env")
    }
    return uri

}

func getDatabase() string {
	var db = os.Getenv("MONGOLAB_DB")
	if db == "" {
        db = "securechat"
        log.Println("INFO: dev_env")
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