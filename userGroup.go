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
	rules map[string]Handler
}

func newUserGroup() *UserGroup {
	self := &UserGroup{}
	self.users = map[*User]bool{}
	self.rules = map[string]Handler{}
	return self
}

// FindHandler implements a handler finding function for router.
func (rt *UserGroup) FindHandler(event string) (Handler, bool) {
	handler, found := rt.rules[event]
	return handler, found
}

// AddEventHandler is a function to add handlers to the router.
func (rt *UserGroup) AddEventHandler(event string, handler Handler) {
	rt.rules[event] = handler
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
}
