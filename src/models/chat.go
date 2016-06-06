package models

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"project/server/src/constants"
	"project/server/src/errorchecker"
	"project/server/src/utils"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
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

// ChatPrivateInfo struct
type ChatPrivateInfo struct {
	Username string
	ChatID   string
	Token    string
}

// ChatToken struct
type ChatToken struct {
	Username string
	Token    string
}

type canal struct {
	chatid string
	conn   []net.Conn
}

//GetChat gets the chat info from the server
func GetChat(id string) Chat {
	var chat Chat

	session, err := mgo.Dial(constants.URI)
	if err != nil {
		fmt.Println("err")
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	usersCollection := session.DB(constants.AuthDatabase).C("chat")
	err = usersCollection.FindId(bson.ObjectIdHex(id)).One(&chat)

	if err != nil {
		fmt.Println("error find chat", err)
	}

	return chat
}

//RecuperarEstado func
func RecuperarEstado(username string) map[string]ChatPrivateInfo {
	var chats []ChatPrivateInfo
	chatsReturn := map[string]ChatPrivateInfo{}
	session, err := mgo.Dial(constants.URI)
	if err != nil {
		fmt.Println("err")
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	chatInfoCollection := session.DB(constants.AuthDatabase).C("chatinfo")
	err = chatInfoCollection.Find(bson.M{"username": username}).All(&chats)
	errorchecker.Check("ERROR buscando chat info", err)

	for _, cht := range chats {
		chatsReturn[cht.ChatID] = cht
		chatInfoCollection.Remove(bson.M{"username": username})
	}
	return chatsReturn
}

//GetChats gets the chats the user has
func GetChats(userid string) []Chat {
	var chats []Chat

	session, err := mgo.Dial(constants.URI)
	if err != nil {
		fmt.Println("err")
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	usersCollection := session.DB(constants.AuthDatabase).C("chat")
	err = usersCollection.Find(bson.M{"components": bson.ObjectIdHex(userid)}).All(&chats)

	errorchecker.Check("ERROR find chat", err)
	return chats
}

//CreateChat creates a chat between sender and receiver
func CreateChat(sender User, receivers []User, chatType string) bson.ObjectId {
	var chat Chat
	var components []bson.ObjectId
	chat.ID = bson.NewObjectId()
	chat.Type = chatType
	chat.Name = sender.Username
	components = append(components, sender.ID)
	for _, receiver := range receivers {
		chat.Name = chat.Name + " y " + receiver.Username
		components = append(components, receiver.ID)
	}
	chat.Components = components
	chat.save()
	return chat.ID
}

var conectados []canal

//OpenChat inits the chat
func OpenChat(connectedUsers map[string]User) {
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
				chatInfo := utils.Decode64(scanner.Text())
				json.Unmarshal(chatInfo, &chat)
			}

			// Get user info
			var user User
			if scanner.Scan() {
				userInfo := utils.Decode64(scanner.Text())
				json.Unmarshal(userInfo, &user)
			}

			connectToChat(chat, conn)

			for scanner.Scan() { // escaneamos la conexión
				text := scanner.Text()

				fmt.Println("cliente[", port, "]: ", text) // mostramos el mensaje del cliente
				for _, conectado := range conectados {
					if conectado.chatid == chat.ID.Hex() {
						for _, c := range conectado.conn {
							if conn != c {
								fmt.Fprintln(c, text)
							}
						}
					}
				}
				chat.NewMessage(user, text)
			}
			disconnectFromChat(chat, conn)
			conn.Close() // cerramos al finalizar el cliente (EOF se envía con ctrl+d o ctrl+z según el sistema)
			fmt.Println("cierre[", port, "]")
		}()
	}
}

func connectToChat(chat Chat, conn net.Conn) {
	existe := false
	for i, conectado := range conectados {
		if conectado.chatid == chat.ID.Hex() {
			conectado.conn = append(conectado.conn, conn)
			existe = true
			conectados[i] = conectado
			break
		}
	}
	if !existe {
		var c canal
		c.chatid = chat.ID.Hex()
		c.conn = append(c.conn, conn)
		conectados = append(conectados, c)
	}

	fmt.Println("conectados", conectados)
}

func disconnectFromChat(chat Chat, conn net.Conn) {
	for i, conectado := range conectados {
		if conectado.chatid == chat.ID.Hex() {
			var connection net.Conn
			for _, c := range conectado.conn {
				if c != conn {
					connection = c
				}
			}
			if connection != nil {
				conectado.conn = []net.Conn{connection}
			} else {
				conectado.conn = []net.Conn{}
			}
			conectados[i] = conectado
			break
		}
	}
	fmt.Println("conectados", conectados)
}

func (c *Chat) save() Chat {
	var chat Chat
	session, err := mgo.Dial(constants.URI)
	errorchecker.Check("ERROR dialing", err)
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	collection := session.DB(constants.AuthDatabase).C("chat")
	err = collection.Insert(&c)
	errorchecker.Check("ERROR inserting chat", err)

	return chat
}

// SaveChatInfo func
func SaveChatInfo(tokens []ChatToken, chatid bson.ObjectId) {
	var infoChats []ChatPrivateInfo
	var chatInfo ChatPrivateInfo

	for _, token := range tokens {
		chatInfo.Token = token.Token
		chatInfo.Username = token.Username
		chatInfo.ChatID = chatid.Hex()
		infoChats = append(infoChats, chatInfo)
	}

	session, err := mgo.Dial(constants.URI)
	errorchecker.Check("ERROR dialing", err)
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	collection := session.DB(constants.AuthDatabase).C("chatinfo")
	for _, chat := range infoChats {
		err = collection.Insert(&chat)
		errorchecker.Check("ERROR inserting chatsInfo", err)
	}
}

//NewMessage adds the message to the chat
func (c *Chat) NewMessage(user User, msg string) {
	session, err := mgo.Dial(constants.URI)
	errorchecker.Check("ERROR dialing", err)
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	collection := session.DB(constants.AuthDatabase).C("chat")
	change := bson.M{"$push": bson.M{"messages": bson.M{"id": bson.NewObjectId(), "content": msg, "sender": user.ID.Hex(), "date": time.Now().String()}}}
	err = collection.UpdateId(c.ID, change)
	errorchecker.Check("ERROR inserting message", err)
}
