package src

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
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
