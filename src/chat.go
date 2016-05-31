package src

import (
	"bufio"
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net"
)

//Chat structure
type Chat struct {
	ID         bson.ObjectId   `bson:"_id,omitempty"`
	Name       string          `bson:"name"`
	Type       string          `bson:"type"`
	Components []bson.ObjectId `bson:"components"`
	Messages   []string        `bson:"messages"`
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
		fmt.Println("en bucle")
		conn, err := ln.Accept() // para cada nueva petición de conexión
		if err != nil {
			fmt.Println("ERROR", err)
		}
		fmt.Println("antes lambda")
		go func() { // lanzamos un cierre (lambda, función anónima) en concurrencia
			fmt.Println("EN lambda")
			_, port, err := net.SplitHostPort(conn.RemoteAddr().String()) // obtenemos el puerto remoto para identificar al cliente (decorativo)
			if err != nil {
				fmt.Println("ERROR", err)
			}

			fmt.Println("conexión: ", conn.LocalAddr(), " <--> ", conn.RemoteAddr())

			scanner := bufio.NewScanner(conn) // el scanner nos permite trabajar con la entrada línea a línea (por defecto)

			for scanner.Scan() { // escaneamos la conexión
				fmt.Println("cliente[", port, "]: ", scanner.Text()) // mostramos el mensaje del cliente
				fmt.Fprintln(conn, "ack: ", scanner.Text())          // enviamos ack al cliente
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
