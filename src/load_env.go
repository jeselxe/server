package src

import (
    "os"
)

var URI = os.Getenv("MONGOLAB_URI")
var AUTH_DATABASE = os.Getenv("MONGOLAB_DB")