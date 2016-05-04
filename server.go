package main

import (
	"encoding/base64"
	"encoding/json"
	"golang.org/x/crypto/scrypt"
	"golang.org/x/net/websocket"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"net/http"
	"project/server/src"
	"strconv"
)

// User structure
type User struct {
	ID        bson.ObjectId `bson:"_id,omitempty"`
	Username  string        `bson:"name"`
	Password  string        `bson:"password"`
	Salt      string        `bson:"salt"`
	PubKey    string        `bson:"pubkey"`
	PrivKey   string        `bson:"privkey"`
	CipherMsg string        `bson:"ciphermsg"`
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

	salt, _ := base64.StdEncoding.DecodeString(user.Salt)

	dk, err := scrypt.Key(passwd, salt, 16384, 8, 1, 32)

	if err != nil {
		return User{}
	}

	if user.Password == base64.StdEncoding.EncodeToString(dk) {
		return user
	}
	return User{}

}

func registerUser(user *User) bool {
	session, err := mgo.Dial(src.URI)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	c := session.DB(src.AUTH_DATABASE).C("user")

	err = c.Insert(&user)

	if err != nil {
		return false
	}

	return true
}

func encrypt(pass []byte) (dk, salt []byte) {
	salt = []byte("random salt")
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
	user.Password = base64.StdEncoding.EncodeToString(pwd)
	user.Salt = base64.StdEncoding.EncodeToString(salt)
	user.PubKey = r.FormValue("pub")
	user.PrivKey = r.FormValue("priv")
	user.CipherMsg = r.FormValue("msg")

	registered := registerUser(&user)

	w.Write([]byte(strconv.AppendBool(make([]byte, 0), registered)))
}

func main() {
	http.Handle("/chat", websocket.Handler(chatHandler))
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/register", registerHandler)

	err := http.ListenAndServe(src.PORT, nil)
	if err != nil {
		log.Fatal("ListenAndServe: " + err.Error())
	}
}
