package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"project/server/src/constants"
	"project/server/src/models"
	"project/server/src/utils"
	"strconv"

	"golang.org/x/net/websocket"
)

// start docs
// godoc -http=:6060
var connectedUsers map[string]models.User

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

func checkLogin(username string, passwd []byte) models.User {
	fmt.Println("Usuario <" + username + "> intenta loguearse.")
	user := models.SearchUser(username)

	if user.Validate() {
		fmt.Println(user.GetSalt())
		salt := utils.Decode64(user.GetSalt())
		hashedPasswd, err := utils.ScryptHash(passwd, salt)
		if err == nil {
			if user.Password == utils.Encode64(hashedPasswd) {
				fmt.Println("Login correcto.")
				addConnectedUser(user)
				return user
			}
		}
	}
	fmt.Println("Login usuario <" + username + "> rechazado.")
	return models.User{}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	// Check user chats and return them
	username := r.FormValue("username")
	pass := utils.Decode64(r.FormValue("pass"))

	user := checkLogin(username, pass)
	res, _ := json.Marshal(user)
	user.Print()
	w.Write(res)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	fmt.Println("Usuario <" + username + "> intenta registrarse.")
	password := r.FormValue("pass")
	pubKey := r.FormValue("pub")
	state := r.FormValue("state")
	user, err := models.RegisterUser(username, password, pubKey, state)
	if err != nil {
		fmt.Println("Usuario <" + username + "> registro rechazado.")
		w.Write([]byte("{error: 'user exists'}"))
	} else {
		fmt.Println("Usuario <" + username + "> registrado correctamente.")
		addConnectedUser(user)
		res, _ := json.Marshal(user)
		w.Write(res)
	}
}

func searchUsersHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	users := models.SearchUsers(username)
	res, _ := json.Marshal(users)
	w.Write(res)
}

func newChatHandler(w http.ResponseWriter, r *http.Request) {
	var user models.User
	senderUsername := r.FormValue("sender")
	senderState := r.FormValue("senderState")

	user = models.SearchUser(senderUsername)

	user.State = senderState
	user.Save()

	receivername := r.FormValue("receiver")
	receiverKey := r.FormValue("receiverkey")

	sender := models.SearchUser(senderUsername)
	receiver := models.SearchUser(receivername)

	chatid := models.CreateChat(sender, receiver)

	models.SaveChatInfo(receiver.Username, receiverKey, chatid)

	receiver.AddChat(chatid, receiverKey)

	var chat models.Chat
	chat = models.GetChat(chatid.Hex())

	res, _ := json.Marshal(chat)
	w.Write(res)
}

func getChatsHandler(w http.ResponseWriter, r *http.Request) {
	userid := r.FormValue("userid")

	chats := models.GetChats(userid)

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

func addConnectedUser(user models.User) {
	connectedUsers[user.Username] = user
	printConnectedUsers()
}

func removeConnectedUser(username string) {
	delete(connectedUsers, username)
	printConnectedUsers()
}

func updateStateHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	fmt.Println("Usuario <" + username + "> intenta recuperar su estado.")
	chatsInfo := models.RecuperarEstado(username)
	byteChats, _ := json.Marshal(chatsInfo)
	byteChats = utils.Compress(byteChats)
	w.Write(byteChats)
}

func main() {
	connectedUsers = make(map[string]models.User)
	http.Handle("/chat", websocket.Handler(chatHandler))
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/search_user", searchUsersHandler)
	http.HandleFunc("/new_chat", newChatHandler)
	http.HandleFunc("/get_chats", getChatsHandler)
	http.HandleFunc("/update_state", updateStateHandler)
	go models.OpenChat(connectedUsers)
	err := http.ListenAndServe(constants.Port, nil)
	if err != nil {
		fmt.Println("ListenAndServe error")
	}
}
