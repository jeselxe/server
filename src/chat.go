package src

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net"
	"time"
)

// Message structure
type Message struct {
	ID      bson.ObjectId `bson:"id"`
	Content string        `bson:"content"`
	Date    string        `bson:"date"`
	Sender  string        `bson:"sender"`
}

//Chat structure
type Chat struct {
	ID         bson.ObjectId   `bson:"_id,omitempty"`
	Name       string          `bson:"name"`
	Type       string          `bson:"type"`
	Components []bson.ObjectId `bson:"components"`
	Messages   []Message       `bson:"messages"`
}

//GetChat gets the chat info from the server
func GetChat(id string) Chat {
	var chat Chat

	session, err := mgo.Dial(URI)
	if err != nil {
		fmt.Println("err")
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	usersCollection := session.DB(AuthDatabase).C("chat")
	err = usersCollection.FindId(bson.ObjectIdHex(id)).One(&chat)

	if err != nil {
		fmt.Println("error find chat", err)
	}

	return chat
}

//GetChats gets the chats the user has
func GetChats(userid string) []Chat {
	var chats []Chat

	session, err := mgo.Dial(URI)
	if err != nil {
		fmt.Println("err")
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	usersCollection := session.DB(AuthDatabase).C("chat")
	err = usersCollection.Find(bson.M{"components": bson.ObjectIdHex(userid)}).All(&chats)

	if err != nil {
		fmt.Println("error find chat", err)
	}

	return chats
}

//CreateChat creates a chat between sender and receiver
func CreateChat(sender, receiver User) bson.ObjectId {
	var chat Chat

	chat.ID = bson.NewObjectId()
	chat.Type = "individual"
	chat.Name = sender.Username + " y " + receiver.Username
	chat.Components = []bson.ObjectId{sender.ID, receiver.ID}

	chat.save()
	return chat.ID
}

//OpenChat inits the chat
func OpenChat() {
	ln, err := net.Listen("tcp", "localhost:1337") // escucha en espera de conexión
	if err != nil {
		fmt.Println("ERROR", err)
	}
	defer ln.Close() // nos aseguramos que cerramos las conexiones aunque el programa falle

	for { // búcle infinito, se sale con ctrl+c

		conn, err := ln.Accept() // para cada nueva petición de conexión
		if err != nil {
			fmt.Println("ERROR", err)
		}

		go func() { // lanzamos un cierre (lambda, función anónima) en concurrencia
			_, port, err := net.SplitHostPort(conn.RemoteAddr().String()) // obtenemos el puerto remoto para identificar al cliente (decorativo)
			if err != nil {
				fmt.Println("ERROR", err)
			}

			fmt.Println("conexión: ", conn.LocalAddr(), " <--> ", conn.RemoteAddr())

			scanner := bufio.NewScanner(conn) // el scanner nos permite trabajar con la entrada línea a línea (por defecto)

			// Get chat info
			var chat Chat
			if scanner.Scan() {
				chatInfo, _ := base64.StdEncoding.DecodeString(scanner.Text())
				json.Unmarshal(chatInfo, &chat)
			}

			// Get user info
			var user User
			if scanner.Scan() {
				userInfo, _ := base64.StdEncoding.DecodeString(scanner.Text())
				json.Unmarshal(userInfo, &user)
			}

			for scanner.Scan() { // escaneamos la conexión
				text := scanner.Text()

				fmt.Println("cliente[", port, "]: ", text) // mostramos el mensaje del cliente
				fmt.Fprintln(conn, scanner.Text())         // enviamos ack al cliente

				chat.NewMessage(user, text)
			}

			conn.Close() // cerramos al finalizar el cliente (EOF se envía con ctrl+d o ctrl+z según el sistema)
			fmt.Println("cierre[", port, "]")
		}()
	}
}

func (c *Chat) save() Chat {
	var chat Chat
	session, err := mgo.Dial(URI)
	Check(err)
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	collection := session.DB(AuthDatabase).C("chat")
	err = collection.Insert(&c)
	Check(err)

	return chat
}

//NewMessage adds the message to the chat
func (c *Chat) NewMessage(user User, msg string) {
	session, err := mgo.Dial(URI)
	Check(err)
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	collection := session.DB(AuthDatabase).C("chat")
	change := bson.M{"$push": bson.M{"messages": bson.M{"id": bson.NewObjectId(), "content": msg, "sender": user.ID.Hex(), "date": time.Now().String()}}}
	err = collection.UpdateId(c.ID, change)
	if err != nil {
		fmt.Println("error new message", err)
	}
}
