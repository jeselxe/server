package src

import (
	"errors"
	"fmt"
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

// Login given a username and password, it tries to return its info from DB
func Login(username string, password []byte) User {
	user := SearchUser(username)[0]
	if user.Validate() {
		salt := Decode64(user.Salt)
		hashedPasswd, err := ScryptHash(password, salt)
		if err == nil {
			if user.Password == Encode64(hashedPasswd) {
				return user
			}
		}
	}
	return User{}
}

// RegisterUser registered
func RegisterUser(username, password, pubKey, privKey string) (User, error) {
	var user User
	var returnError error
	users := SearchUser(username)
	if len(users) == 1 {
		returnError = errors.New("Username taken")
	} else {
		decodedPassword := Decode64(password)
		hashedPassword, salt := HashWithRandomSalt(decodedPassword)

		user.Username = username
		user.Password = Encode64(hashedPassword)
		user.Salt = Encode64(salt)
		user.PubKey = pubKey
		user.PrivKey = privKey

		user.Print()

		user = user.save()
	}
	return user, returnError
}

// Print prints invoking user
func (u *User) Print() {
	fmt.Println("################### USER #####################")
	fmt.Println(u.ID)
	fmt.Println(u.Username)
	fmt.Println(u.Password)
	fmt.Println(u.PrivKey)
	fmt.Println(u.PubKey)
	fmt.Println(u.Salt)
	fmt.Println("################# END USER ###################")
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

	users := SearchUser(u.Username)
	if len(users) == 1 {
		user = users[0]
	}

	return user
}

// SearchUser returns the User object given a user with username
func SearchUser(username string) []User {
	var user User
	var users []User

	session, err := mgo.Dial(URI)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	usersCollection := session.DB(AuthDatabase).C("user")
	err = usersCollection.Find(bson.M{"name": username}).One(&user)
	if !user.Validate() {
		username = "/" + username + "/"
		fmt.Println(username)
		// FIXME error en query
		err = usersCollection.Find(bson.M{"name": bson.RegEx{"/a/", ""}}).All(&users)
		fmt.Println(len(users))
	} else {
		users = append(users, user)
	}

	if err != nil {
		log.Println("error count", err)
	}

	return users
}

// Validate given a user u it returns whether its attributes are valid or not
func (u *User) Validate() bool {
	if u.Username == "" {
		return false
	}
	return true
}
