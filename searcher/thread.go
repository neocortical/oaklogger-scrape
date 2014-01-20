package searcher

import (
	"time"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

type Thread struct {
	Id						bson.ObjectId `json:"id"        bson:"_id,omitempty"`
	TID						int           `json:"tid"`
	Name 					string				`json:"name"`
	UID 					int						`json:"uid"`
	PostCount			int						`json:"post_count"`
	LastPostTime 	time.Time     `json:"last_post_time"`
	HasPosts      bool          `json:"has_posts"`
	HasTwin       bool          `json:"has_twin"`
}

type threadSearcher struct {
	s *mgo.Session
	c *mgo.Collection
	closed bool
}

func NewThreadSearcher() (*threadSearcher) {
	ts := new(threadSearcher)
	
	session, err := mgo.Dial("localhost:27017")
	if err != nil {
		panic(err)
	}
	ts.s = session
	
	session.SetMode(mgo.Monotonic, true)
	
	ts.c = session.DB("oaklogger").C("threads")
	ts.closed = false
	
	return ts
}

func (ts *threadSearcher) FindThread(name string) (*Thread) {
	if ts.closed {
		panic("This searcher has been closed")
	}

	var result Thread
	err := ts.c.Find(bson.M{"name": name, "hasposts": true}).One(&result)
	if err != nil {
	    return nil
	}
	
	return &result
}

/** Hack fix for Oaklog bug where some threads' messages only show the first word of the thread title. */
func (ts *threadSearcher) FindThreadByFirstWord(name string) (*Thread) {
	if ts.closed {
		panic("This searcher has been closed")
	}

	var result Thread
	err := ts.c.Find(bson.M{"name": bson.RegEx{"^"+name+"\\s.*", ""}, "hasposts": true}).One(&result)
	if err != nil {
	    return nil
	}
	
	return &result
}

func (ts *threadSearcher) Save(t *Thread) {
	if ts.closed {
		panic("This searcher has been closed")
	}

	_, err := ts.c.Upsert(bson.M{"tid": t.TID}, t)
	if err != nil {
	    panic("unable to save to DB")
	}
}

func (ts *threadSearcher) Close() {
	if !ts.closed {
		ts.s.Close()
		ts.closed = true
	}
}