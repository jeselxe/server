package main

import (
    "log"
    "net/http"
    "os"
    "golang.org/x/net/websocket"
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
    "secure-chat/src"
    "strconv"
)

// Get the Port from the environment so we can run on Heroku
func GetPort() string {
    var port = os.Getenv("PORT")
    // Set a default port if there is nothing in the environment
    if port == "" {
        port = "8080"
        log.Println("INFO: No PORT environment variable detected, defaulting to " + port)
    }
    return ":" + port
}


/*
// User structure
type User struct {
    ID       bson.ObjectId `bson:"_id,omitempty"`
    Username     string        `bson:"username"`
    Password string        `bson:"password"`
}
*/
// Message structure
type Message struct {
    ID       bson.ObjectId `bson:"_id,omitempty"`
    Content     string        `bson:"content"`
}


func mongo(message string) string{
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
    if err != nil{
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

func main() {
    http.Handle("/chat", websocket.Handler(chatHandler))
    port := GetPort()
    err := http.ListenAndServe(port, nil)
    if err != nil {
        log.Fatal("ListenAndServe: " + err.Error())
    }
}
