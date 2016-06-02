package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"project/client/src/utils"
	"project/server/src"
	"strconv"

	"golang.org/x/net/websocket"
)

// start docs
// godoc -http=:6060
var connectedUsers map[string]src.User

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
	fmt.Println("Usuario <" + username + "> intenta loguearse.")
	users := src.SearchUser(username)

	if len(users) == 1 {
		user := users[0]
		salt := src.Decode64(user.Salt)
		hashedPasswd, err := src.ScryptHash(passwd, salt)
		if err == nil {
			if user.Password == src.Encode64(hashedPasswd) {
				fmt.Println("Login correcto.")
				addConnectedUser(user)
				return user
			}
		}
	}
	fmt.Println("Login rechazado.")
	return src.User{}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	// Check user chats and return them
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
	username := r.FormValue("username")
	users := src.SearchUser(username)
	res, _ := json.Marshal(users)
	w.Write(res)
}

func newChatHandler(w http.ResponseWriter, r *http.Request) {
	sendername := r.FormValue("sender")
	senderKey := r.FormValue("senderkey")
	receivername := r.FormValue("receiver")
	receiverKey := r.FormValue("receiverkey")

	users := src.SearchUser(sendername)
	sender := users[0]
	users = src.SearchUser(receivername)
	receiver := users[0]

	chatid := src.CreateChat(sender, receiver)

	sender.AddChat(chatid, senderKey)
	receiver.AddChat(chatid, receiverKey)

	var chat src.Chat
	chat = src.GetChat(chatid.Hex())

	res, _ := json.Marshal(chat)
	w.Write(res)
}

func getChatsHandler(w http.ResponseWriter, r *http.Request) {
	userid := r.FormValue("userid")

	chats := src.GetChats(userid)

	res, _ := json.Marshal(chats)
	w.Write(res)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	fmt.Println("Usuario <" + username + "> intenta hacer logout.")
	removeConnectedUser(username)
	fmt.Println("Logout correcto")
	w.Write([]byte("Logout"))
}

func printConnectedUsers() {
	index := 1
	fmt.Println("Hay " + strconv.Itoa(len(connectedUsers)) + " usuarios conectados:")
	for key := range connectedUsers {
		fmt.Println(strconv.Itoa(index) + ": " + key)
		index++
	}
}

func addConnectedUser(user src.User) {
	connectedUsers[user.Username] = user
	printConnectedUsers()
}

func removeConnectedUser(username string) {
	delete(connectedUsers, username)
	printConnectedUsers()
}

func main() {
	connectedUsers = make(map[string]src.User)
	http.Handle("/chat", websocket.Handler(chatHandler))
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/search_user", searchUserHandler)
	http.HandleFunc("/new_chat", newChatHandler)
	http.HandleFunc("/get_chats", getChatsHandler)

	go src.OpenChat()

	err := http.ListenAndServe(src.Port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: " + err.Error())
	}
}
