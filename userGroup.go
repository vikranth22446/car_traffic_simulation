package main

import (
	"github.com/google/uuid"
	"math/rand"
	"os"
	"time"
)

type UserGroup struct {
	users map[*User]bool
	ids   map[uuid.UUID]*User
}

func newUserGroup() *UserGroup {
	self := &UserGroup{}
	self.users = map[*User]bool{}
	return self
}

func (self *UserGroup) removePlayer(p *User) {
	delete(self.users, p)
	delete(self.ids, p.Id)
	// close(p.send)
	p.group = nil
}
var (
	userGroup *UserGroup
)

func init() {
	userGroup = newUserGroup()
	if os.Getenv("DEBUG") != "" {
		playerShowLog = true
	}
	rand.Seed(time.Now().Unix())
}

func (self *UserGroup) addUser(p *User) {
	//var chunk []byte

	p.group = self
	newUUid, err := uuid.NewUUID()
	//(err)
	if err != nil {
		return
	}
	p.Id = newUUid
	self.users[p] = true

	//
	//if p.ws == nil {
	//	chunk = createFn("ship-ai", p.Id, p.Serialize())
	//} else {
	//	chunk = createFn("ship", p.Id, p.Serialize())
	//}
	//
	//// Announcing new player
	//self.broadcast(chunk)
	//
	//// Announcing existing elements.
	//for other, _ := range self.players {
	//	if p.sameAs(other) == false {
	//		p.notice(other)
	//	}
	//}

}