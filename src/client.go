package src

import (
    "fmt"
    "io"
    "log"
    "golang.org/x/net/websocket"
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
)

type Client struct {
    id     bson.ObjectId
    Username string
    Password string
    ws     *websocket.Conn
    server *Server
}

// ClientDB structure
type ClientDB struct {
    ID       bson.ObjectId `bson:"_id,omitempty"`
    Username string        `bson:"username"`
    Password string        `bson:"password"`
}

// Create new chat client.
func NewClient(ws *websocket.Conn, server *Server) *Client {
    if ws == nil {
        log.Fatal("ws cannot be nil")
    }

    if server == nil {
        log.Fatal("server cannot be nil")
    }

    return &Client{ws, server, ch, doneCh}

//
func (c *Client) Login(client Client) Bool {
    
    user, err = c.Find(bson.M{"username": client.Username, "password": client.Password}).One(&result)
    if err != nil {
        log.Println("User not found")
        log.Fatal(err)
    }
}