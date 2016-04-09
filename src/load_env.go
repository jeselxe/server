package src

import (
	"os"
)

var develop = os.Getenv("PROD")

var URI = getURI()
var AUTH_DATABASE = getDatabase()

func getURI() string {
	if develop == "true" {
		return os.Getenv("MONGOLAB_URI")
	}
	return "localhost:27017"

}

func getDatabase() string {
	if develop == "true" {
		return os.Getenv("MONGOLAB_DB")
	}
	return "securechat"
}
