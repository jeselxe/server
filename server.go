package main

import (
	"golang.org/x/net/websocket"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"net/http"
	"os"
	"project/server/src"
	"strconv"
)

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

// User structure
type User struct {
	ID       bson.ObjectId `bson:"_id,omitempty"`
	Username string        `bson:"name"`
	Password string        `bson:"password"`
}

// Message structure
type Message struct {
	ID      bson.ObjectId `bson:"_id,omitempty"`
	Content string        `bson:"content"`
}

func mongo(message string) string {
	session, err := mgo.Dial(src.URI)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	// Collection messages
	c := session.DB(src.AUTH_DATABASE).C("messages")

	// Insert
	result := Message{}
	err = c.Insert(&Message{Content: message})
	if err != nil {
		log.Fatal(err)
	}

	// Query count
	messages, err := c.Count()
	// Query
	//err = c.Find(bson.M{"username": "david"}) .One(&result)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(result)

	// fmt.Println("Id: ", result.ID, " - User: ", result.Name, " - Password: ", result.Password)
	response := "Total messages =" + strconv.Itoa(messages)
	return response
}

func chatHandler(ws *websocket.Conn) {
	msg := make([]byte, 512)
	n, err := ws.Read(msg)
	if err != nil {
		log.Fatal(err)
	}
	message := string(msg[:n])
	log.Printf("Receive: %s\n", message)

	response := mongo(message)

	_, err = ws.Write([]byte(response))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Send: %s\n", response)
}

func searchUser(user, passwd string) bool {
	session, err := mgo.Dial("localhost:27017")
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	c := session.DB("securechat").C("user")

	result := User{}

	err = c.Find(bson.M{"name": user, "password": passwd}).One(&result)

	if err != nil {
		log.Println("error count", err)
		return false
	}

	return result.ID.Valid()
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Request: ", r.FormValue("user"), r.FormValue("pass"))
	logged := searchUser(r.FormValue("user"), r.FormValue("pass"))

	w.Write([]byte(strconv.AppendBool(make([]byte, 0), logged)))
}

func main() {
	http.Handle("/chat", websocket.Handler(chatHandler))
	http.HandleFunc("/login", loginHandler)
	port := GetPort()
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: " + err.Error())
	}
}
