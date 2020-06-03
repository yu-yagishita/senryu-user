package mongodb

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/yu-yagishita/nanpa-user/users"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	name     string
	password string
	host     string
	db       = "users"
	//ErrInvalidHexID represents a entity id that is not a valid bson ObjectID
	ErrInvalidHexID = errors.New("Invalid Id Hex")
)

func init() {
	fmt.Println("name:" + os.Getenv("MONGO_USER"))
	fmt.Println("password:" + os.Getenv("MONGO_PASS"))
	fmt.Println("host:" + os.Getenv("MONGO_HOST"))
	flag.StringVar(&name, "mongo-user", os.Getenv("MONGO_USER"), "Mongo user")
	flag.StringVar(&password, "mongo-password", os.Getenv("MONGO_PASS"), "Mongo password")
	flag.StringVar(&host, "mongo-host", os.Getenv("MONGO_HOST"), "Mongo host")
}

// Mongo meets the Database interface requirements
type Mongo struct {
	//Session is a MongoDB Session
	Session *mgo.Session
}

// Init MongoDB
func (m *Mongo) Init() error {
	fmt.Println("MongoDB: Init")
	u := getURL()
	fmt.Println("u: " + u.String())
	var err error
	m.Session, err = mgo.DialWithTimeout(u.String(), time.Duration(5)*time.Second)
	if err != nil {
		return err
	}
	return m.EnsureIndexes()
}

// MongoUser is a wrapper for the users
type MongoUser struct {
	users.User `bson:",inline"`
	ID         bson.ObjectId `bson:"_id"`
	// AddressIDs []bson.ObjectId `bson:"addresses"`
	// CardIDs    []bson.ObjectId `bson:"cards"`
}

// New Returns a new MongoUser
func New() MongoUser {
	u := users.New()
	return MongoUser{
		User: u,
		// AddressIDs: make([]bson.ObjectId, 0),
		// CardIDs:    make([]bson.ObjectId, 0),
	}
}

// CreateUser Insert user to MongoDB, including connected addresses and cards, update passed in user with Ids
func (m *Mongo) CreateUser(u *users.User) error {
	s := m.Session.Copy()
	defer s.Close()
	id := bson.NewObjectId()
	mu := New()
	mu.User = *u
	mu.ID = id
	// var carderr error
	// var addrerr error
	// mu.CardIDs, carderr = m.createCards(u.Cards)
	// mu.AddressIDs, addrerr = m.createAddresses(u.Addresses)
	c := s.DB("").C("users")
	_, err := c.UpsertId(mu.ID, mu)
	if err != nil {
		// Gonna clean up if we can, ignore error
		// because the user save error takes precedence.
		// m.cleanAttributes(mu)
		return err
	}
	mu.User.UserID = mu.ID.Hex()
	// Cheap err for attributes
	// if carderr != nil || addrerr != nil {
	// 	return fmt.Errorf("%v %v", carderr, addrerr)
	// }
	*u = mu.User
	return nil
}

// GetUserByName Get user by their name
func (m *Mongo) GetUserByName(name string) (users.User, error) {
	fmt.Println("mongodb: GetUserByName")
	s := m.Session.Copy()
	defer s.Close()
	c := s.DB("").C("users")
	fmt.Println("c.Name: " + c.Name)
	mu := New()
	err := c.Find(bson.M{"username": name}).One(&mu)
	fmt.Println("mu.FirstName: " + mu.FirstName)
	// fmt.Println(c.Find(bson.M{"username": name}).One(&mu))
	// mu.AddUserIDs()
	return mu.User, err
}

// // GetUser Get user by their object id
// func (m *Mongo) GetUser(id string) (users.User, error) {
// 	s := m.Session.Copy()
// 	defer s.Close()
// 	if !bson.IsObjectIdHex(id) {
// 		return users.New(), errors.New("Invalid Id Hex")
// 	}
// 	c := s.DB("").C("customers")
// 	mu := New()
// 	err := c.FindId(bson.ObjectIdHex(id)).One(&mu)
// 	// mu.AddUserIDs()
// 	return mu.User, err
// }

// // GetUsers Get all users
// func (m *Mongo) GetUsers() ([]users.User, error) {
// 	// TODO: add paginations
// 	s := m.Session.Copy()
// 	defer s.Close()
// 	c := s.DB("").C("customers")
// 	var mus []MongoUser
// 	err := c.Find(nil).All(&mus)
// 	us := make([]users.User, 0)
// 	for _, mu := range mus {
// 		// mu.AddUserIDs()
// 		us = append(us, mu.User)
// 	}
// 	return us, err
// }

// // CreateAddress Inserts Address into MongoDB
// func (m *Mongo) Delete(entity, id string) error {
// 	if !bson.IsObjectIdHex(id) {
// 		return errors.New("Invalid Id Hex")
// 	}
// 	s := m.Session.Copy()
// 	defer s.Close()
// 	c := s.DB("").C(entity)
// 	if entity == "customers" {
// 		u, err := m.GetUser(id)
// 		if err != nil {
// 			return err
// 		}
// 		// // aids := make([]bson.ObjectId, 0)
// 		// // for _, a := range u.Addresses {
// 		// // 	aids = append(aids, bson.ObjectIdHex(a.ID))
// 		// // }
// 		// // cids := make([]bson.ObjectId, 0)
// 		// // for _, c := range u.Cards {
// 		// // 	cids = append(cids, bson.ObjectIdHex(c.ID))
// 		// // }
// 		// ac := s.DB("").C("addresses")
// 		// ac.RemoveAll(bson.M{"_id": bson.M{"$in": aids}})
// 		// cc := s.DB("").C("cards")
// 		// cc.RemoveAll(bson.M{"_id": bson.M{"$in": cids}})
// 	// } else {
// 		// c := s.DB("").C("customers")
// 		// c.UpdateAll(bson.M{},
// 		// 	bson.M{"$pull": bson.M{entity: bson.ObjectIdHex(id)}})
// 	// }
// 	return c.Remove(bson.M{"_id": bson.ObjectIdHex(id)})
// }

func getURL() url.URL {
	ur := url.URL{
		Scheme: "mongodb",
		Host:   host,
		Path:   db,
	}
	if name != "" {
		u := url.UserPassword(name, password)
		ur.User = u
	}
	return ur
}

// EnsureIndexes ensures username is unique
func (m *Mongo) EnsureIndexes() error {
	s := m.Session.Copy()
	defer s.Close()
	i := mgo.Index{
		Key:        []string{"username"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     false,
	}
	c := s.DB("").C("users")
	return c.EnsureIndex(i)
}

// func (m *Mongo) Ping() error {
// 	s := m.Session.Copy()
// 	defer s.Close()
// 	return s.Ping()
// }
