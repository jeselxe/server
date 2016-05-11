package src

import (
	"log"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// User structure
type User struct {
	ID       bson.ObjectId `bson:"_id,omitempty"`
	Username string        `bson:"name"`
	Password string        `bson:"password"`
	Salt     string        `bson:"salt"`
	PubKey   string        `bson:"pubkey"`
	PrivKey  string        `bson:"privkey"`
}

// Search searchs a user with the given username
func (u *User) Search() User {
	var user User

	session, err := mgo.Dial(URI)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	users := session.DB(AuthDatabase).C("user")

	err = users.Find(bson.M{"name": u.Username}).One(&user)
	Check(err)

	return user
}

// Login given a user, it tries to return its info from DB
func (u *User) Login() User {
	var user User
	user = u.Search()

	if user.Username != "" {
		/*
			salt := Decode64(user.Salt)
			//hashedPasswd, err := Scrypt(passwd, salt)
				if err == nil {
					if user.Password == Encode64(hashedPasswd) {
						return user
					}
				}
		*/
	}
	return User{}
}

// RegisterUser registered
func RegisterUser(username, password, pubKey, privKey string) User {
	var user User

	user = SearchUser(username)
	if user.ID == "" {
		decodedPassword := Decode64(password)
		hashedPassword, salt := HashWithRandomSalt(decodedPassword)

		user.Username = username
		user.Password = Encode64(hashedPassword)
		user.Salt = Encode64(salt)
		user.PubKey = pubKey
		user.PrivKey = privKey

		user = user.save()
	}

	return user
}

func (u *User) save() User {
	var user User
	session, err := mgo.Dial(URI)
	Check(err)
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	c := session.DB(AuthDatabase).C("user")
	err = c.Insert(&u)
	Check(err)

	user = SearchUser(u.Username)

	return user
}

// SearchUser returns the User object given a user with username
func SearchUser(username string) User {
	var user User

	session, err := mgo.Dial(URI)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	users := session.DB(AuthDatabase).C("user")

	err = users.Find(bson.M{"name": username}).One(&user)

	if err != nil {
		log.Println("error count", err)
	}

	return user
}
