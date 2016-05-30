package main

import (
	"encoding/json"
	"log"
	"net/http"
	"project/client/src/utils"
	"project/server/src"

	"golang.org/x/net/websocket"
	"gopkg.in/mgo.v2/bson"
)

// start docs
// godoc -http=:6060

// Message structure
type Message struct {
	ID      bson.ObjectId `bson:"_id,omitempty"`
	Content string        `bson:"content"`
}

func chatHandler(ws *websocket.Conn) {
	msg := make([]byte, 512)
	n, err := ws.Read(msg)
	if err != nil {
		log.Fatal(err)
	}
	message := string(msg[:n])
	log.Printf("Receive: %s\n", message)

	response := ""

	_, err = ws.Write([]byte(response))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Send: %s\n", response)
}

func checkLogin(username string, passwd []byte) src.User {
	var user src.User
	user = src.SearchUser(username)
	salt := src.Decode64(user.Salt)
	hashedPasswd, err := src.ScryptHash(passwd, salt)

	if err == nil {
		if user.Password == src.Encode64(hashedPasswd) {
			return user
		}
	}

	return src.User{}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	pass := utils.Decode64(r.FormValue("pass"))

	user := checkLogin(username, pass)
	res, _ := json.Marshal(user)
	w.Write(res)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("pass")
	pubKey := r.FormValue("pub")
	privKey := r.FormValue("priv")

	user, err := src.RegisterUser(username, password, pubKey, privKey)
	if err != nil {
		w.Write([]byte("{error: 'user exists'}"))
	} else {
		res, _ := json.Marshal(user)
		w.Write(res)
	}
}

func searchUserHandler(w http.ResponseWriter, r *http.Request) {
	var user src.User
	username := r.FormValue("username")

	user = src.SearchUser(username)

	res, _ := json.Marshal(user)
	w.Write(res)
}

func main() {
	http.Handle("/chat", websocket.Handler(chatHandler))
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/search_user", searchUserHandler)

	err := http.ListenAndServe(src.Port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: " + err.Error())
	}
}
