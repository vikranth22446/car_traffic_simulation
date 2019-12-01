package main

import (
	"github.com/google/uuid"
	"math/rand"
	"os"
	"time"
)

// UserGroup keeps track of the current weboscket connections
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
func (userGroup *UserGroup) FindHandler(event string) (Handler, bool) {
	handler, found := userGroup.rules[event]
	return handler, found
}

// AddEventHandler is a function to add handlers to the router.
func (userGroup *UserGroup) AddEventHandler(event string, handler Handler) {
	userGroup.rules[event] = handler
}

func (userGroup *UserGroup) removePlayer(p *User) {
	delete(userGroup.users, p)
	delete(userGroup.ids, p.ID)
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

func (userGroup *UserGroup) addUser(p *User) {
	//var chunk []byte

	p.group = userGroup
	newUUID, err := uuid.NewUUID()
	if err != nil {
		return
	}
	p.ID = newUUID
	userGroup.users[p] = true
}
