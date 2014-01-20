package searcher

import (
	"fmt"
	"time"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

type Post struct {
	Id				  bson.ObjectId `json:"id"           bson:"_id,omitempty"`
	PID 			  int						`json:"pid"`
  TID 			  int						`json:"tid"`
  UID 			  int						`json:"uid"`
  Order       int 					`json:"order"`
  Message     string        `json:"message"`
  PostTime    time.Time     `json:"post_time"`
  Edited      bool          `json:"edited"`
  Orphan      bool          `json:"orphan"       bson:",omitempty"`
}

type Status struct {
	PID 			int			`json:"pid"`
	Status    string  `json:"status"`
}

type postSearcher struct {
	s *mgo.Session
	c *mgo.Collection
	metaC *mgo.Collection
	closed bool
}

func NewPostSearcher() (*postSearcher) {
	ps := new(postSearcher)
	
	session, err := mgo.Dial("localhost:27017")
	if err != nil {
		panic(err)
	}
	ps.s = session
	
	session.SetMode(mgo.Monotonic, true)
	
	ps.c = session.DB("oaklogger").C("posts")
	ps.metaC = session.DB("oaklogger").C("status")
	ps.closed = false
	
	return ps
}

func (ps *postSearcher) GetCurrentStatus() (int) {
	if ps.closed {
		panic("This searcher has been closed")
	}
	
	var result Status
	err := ps.metaC.Find(bson.M{"status": "success"}).Sort("-pid").One(&result)
	if err != nil {
	    panic("could not determine scraper status")
	}
	
	return result.PID
}

func (ps *postSearcher) UpdateStatus(pid int, status string) {
	if ps.closed {
		panic("This searcher has been closed")
	}
	
	s := Status{pid, status}
	_, err := ps.metaC.Upsert(bson.M{"pid": pid}, s)
	if err != nil {
		fmt.Println(err)	
	  panic("unable to update status to DB")
	}
}

func (ps *postSearcher) Save(p *Post) {
	if ps.closed {
		panic("This searcher has been closed")
	}

	_, err := ps.c.Upsert(bson.M{"pid": p.PID}, p)
	if err != nil {
		fmt.Println(err)	
	  panic("unable to save to DB")
	}
}

func (ps *postSearcher) Close() {
	if !ps.closed {
		ps.s.Close()
		ps.closed = true
	}
}