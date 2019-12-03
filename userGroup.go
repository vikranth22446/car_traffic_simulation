package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"math/rand"
	"os"
	"strconv"
	"time"
)

// UserGroup keeps track of the current weboscket connections
type UserGroup struct {
	connUserMap map[*websocket.Conn]*User
	users       map[*User]bool
	ids         map[uuid.UUID]*User
	rules       map[string]Handler
}

func newUserGroup() *UserGroup {
	self := &UserGroup{}
	self.users = map[*User]bool{}
	self.rules = map[string]Handler{}
	self.connUserMap = map[*websocket.Conn]*User{}
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
	delete(userGroup.connUserMap, p.ws)
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
	userGroup.AddEventHandler("startSimulation", startSimulationEvent)
	rand.Seed(time.Now().Unix())
}

func (userGroup *UserGroup) addUser(p *User) {
	p.group = userGroup
	newUUID, err := uuid.NewUUID()
	if err != nil {
		return
	}
	p.ID = newUUID
	userGroup.users[p] = true
	userGroup.connUserMap[p.ws] = p
}

func startSimulationEvent(conn *websocket.Conn, data interface{}) {
	fmt.Println("start Simulation event handler")

	user, exists := userGroup.connUserMap[conn]
	if exists {
		config := DefaultGeneralLaneConfig()

		m := data.(map[string]interface{})["data"].(map[string]interface{})
		//fmt.Println(m)
		if sizeOfLane, ok := m["sizeOfLane"].(string); ok {
			sizeOfLaneInt, err := strconv.Atoi(sizeOfLane)
			if err != nil {
				return
			}
			config.sizeOfLane = sizeOfLaneInt
		}

		if numHorizontalLanes, ok := m["numHorizontalLanes"].(string); ok {
			numHorizontalLanesInt, err := strconv.Atoi(numHorizontalLanes)
			if err != nil {
				return
			}
			config.numHorizontalLanes = numHorizontalLanesInt
		}

		if numVerticalLanes, ok := m["numVerticalLanes"].(string); ok {
			numVerticalLanesInt, err := strconv.Atoi(numVerticalLanes)
			if err != nil {
				return
			}
			config.numVerticalLanes = numVerticalLanesInt
		}
		user.runSimulation(config)
	}
}
