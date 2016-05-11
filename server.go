package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"project/server/src"

	"golang.org/x/crypto/scrypt"
	"golang.org/x/net/websocket"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// User structure
type User struct {
	ID       bson.ObjectId `bson:"_id,omitempty"`
	Username string        `bson:"name"`
	password string        `bson:"password"`
	salt     string        `bson:"salt"`
	PubKey   string        `bson:"pubkey"`
	PrivKey  string        `bson:"privkey"`
}

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

func searchUser(username string, passwd []byte) User {
	session, err := mgo.Dial(src.URI)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	c := session.DB(src.AUTH_DATABASE).C("user")

	user := User{}

	err = c.Find(bson.M{"name": username}).One(&user)

	if err != nil {
		log.Println("error count", err)
		return User{}
	}

	salt, _ := base64.StdEncoding.DecodeString(user.salt)

	dk, err := scrypt.Key(passwd, salt, 16384, 8, 1, 32)

	if err != nil {
		return User{}
	}

	if user.password == base64.StdEncoding.EncodeToString(dk) {
		return user
	}
	return User{}

}

func registerUser(inputUser *User) (User, error) {
	var outputUser User
	session, err := mgo.Dial(src.URI)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	c := session.DB(src.AUTH_DATABASE).C("user")
	err = c.Insert(&inputUser)

	if err != nil {
		err = c.Find(bson.M{"name": inputUser.Username}).One(&outputUser)
		if err != nil {
			return outputUser, nil
		}
		return outputUser, errors.New("User not found")
	}

	return outputUser, errors.New("User not inserted")
}

func encrypt(pass []byte) (dk, salt []byte) {
	salt = make([]byte, 16)

	rand.Read(salt)
	dk, err := scrypt.Key(pass, salt, 16384, 8, 1, 32)

	if err != nil {
		log.Println("ERROR SCRYPT", err)
	}
	return
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	pass, _ := base64.StdEncoding.DecodeString(r.FormValue("pass"))

	user := searchUser(r.FormValue("username"), pass)

	//w.Write([]byte(strconv.AppendBool(make([]byte, 0), logged)))

	res, _ := json.Marshal(user)

	w.Write(res)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	pass, _ := base64.StdEncoding.DecodeString(r.FormValue("pass"))

	pwd, salt := encrypt(pass)

	user := User{}
	user.Username = r.FormValue("username")
	user.password = base64.StdEncoding.EncodeToString(pwd)
	user.salt = base64.StdEncoding.EncodeToString(salt)
	user.PubKey = r.FormValue("pub")
	user.PrivKey = r.FormValue("priv")

	registeredUser, _ := registerUser(&user)

	res, _ := json.Marshal(registeredUser)
	w.Write(res)
}

func searchUserHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	username := r.FormValue("username")

	user = findUser(username)

	res, _ := json.Marshal(user)
	w.Write(res)
}

func findUser(username string) User {
	var user User

	collection := dbUsers()

	err := collection.Find(bson.M{"name": username}).One(&user)

	if err != nil {
		log.Println("error count", err)
	}

	return user
}

func dbUsers() *mgo.Collection {
	session, err := mgo.Dial(src.URI)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	collection := session.DB(src.AUTH_DATABASE).C("user")

	return collection
}

func main() {
	http.Handle("/chat", websocket.Handler(chatHandler))
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/search_user", searchUserHandler)

	err := http.ListenAndServe(src.PORT, nil)
	if err != nil {
		log.Fatal("ListenAndServe: " + err.Error())
	}
}
