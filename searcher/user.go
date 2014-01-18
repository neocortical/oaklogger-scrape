package searcher

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

type User struct {
	Id 				bson.ObjectId `json:"id"        bson:"_id,omitempty"`
  UID 			int						`json:"uid"`
  Username	string				`json:"username"`
}

type userSearcher struct {
	s *mgo.Session
	c *mgo.Collection
	closed bool
}

func NewUserSearcher() (*userSearcher) {
	us := new(userSearcher)
	
	session, err := mgo.Dial("localhost:27017")
	if err != nil {
		panic(err)
	}
	us.s = session
	
	session.SetMode(mgo.Monotonic, true)
	
	us.c = session.DB("oaklogger").C("users")
	us.closed = false
	
	return us
}

func (us *userSearcher) FindUser(uid int) (*User) {
	if us.closed {
		panic("This searcher has been closed")
	}

	var result User
	err := us.c.Find(bson.M{"uid": uid}).One(&result)
	if err != nil {
	    return nil
	}
	
	return &result
}

func (us *userSearcher) Save(u *User) {
	if us.closed {
		panic("This searcher has been closed")
	}

	_, err := us.c.Upsert(bson.M{"uid": u.UID}, u)
	if err != nil {
	    panic("unable to save to DB")
	}
}

func (us *userSearcher) Close() {
	if !us.closed {
		us.s.Close()
		us.closed = true
	}
}