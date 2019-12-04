package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
	"net"
	"time"
)

var playerShowLog = false

var (
	// ErrNoWebsocket is the error when websocket fails
	ErrNoWebsocket = errors.New(`don't have any websocket connection`)
)

// User object tracks the current connection and the current simulation
type User struct {
	ws     *websocket.Conn
	output chan []byte
	addr   net.Addr

	group             *UserGroup
	ID                uuid.UUID
	runningSimulation bool
	simulation        *GeneralLaneSimulation
}

// Message is the struct to communicate to the client
type Message struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

// Handler handles messages back from the client
type Handler func(*websocket.Conn, interface{})

func newUser(ws *websocket.Conn) *User {
	self := &User{}
	self.ws = ws

	// Bot players do not use a websocket.
	if self.ws != nil {
		self.output = make(chan []byte, 256)
		//self.compressor = backstream.NewWriter(self, 0)
		self.addr = ws.RemoteAddr()
	} else {
		self.output = nil
	}
	return self
}

func (user *User) log(s string) {
	log.Printf("%s: %s\n", user.addr, s)
}

func (user *User) write(data []byte) {
	if user.ws == nil {
		return
	}
	if user.output == nil {
		return
	}
	user.output <- data
}

// Is function checks if two users are the same
func (user *User) Is(other *User) bool {
	return user.ID == other.ID
}

func (user *User) identify() (error) {
	message := Message{Event: identify, Data: user.ID.String()}
	marshalledMessage, err := json.Marshal(message)
	if err != nil {
		return err
		// TODO handle marshall err
	}
	user.write(marshalledMessage)
	return nil
}

func (user *User) runSimulation(config GeneralLaneSimulationConfig) {
	if user.runningSimulation {
		return
	}
	user.runningSimulation = true

	simulation, err := initMultiLaneSimulation(config)
	if err != nil {
		// TODO handle this
	}
	user.simulation = simulation
	user.sendUpdatedSimulation()

	go RunGeneralSimulation(simulation)
	for {
		if !simulation.runningSimulation {
			fmt.Println("General Lane Simulation completed")
			if user.runningSimulation {
				user.sendCompletedSimulation() // only send completed if already running
				user.runningSimulation = false
			}
			return
		}
		select {
		case <-simulation.drawUpdateChan:
			user.sendUpdatedSimulation()
			fmt.Println(user.simulation)
			break
		}
	}
}

//
//func (user *User) runSingleSimulation() {
//	if user.runningSimulation {
//		return
//	}
//	user.runningSimulation = true
//
//	simulation := initSingleLaneSimulation(10)
//	user.simulation = simulation
//
//	go RunSingleLaneSimulation(simulation)
//	for {
//		if !simulation.runningSimulation {
//			fmt.Println("SingleLaneSimulation completed")
//			return
//		}
//		select {
//		case <-simulation.drawUpdateChan:
//			user.sendUpdatedSimulation()
//			break
//		}
//	}
//}

func (user *User) sendCompletedSimulation() (error) {
	message := Message{Event: completedSimulation, Data: "Completed"}
	marshalledMessage, err := json.Marshal(message)
	if err != nil {
		panic(err)
		return err
		// TODO handle marshall err
	}
	user.write(marshalledMessage)
	return nil
}

func (user *User) sendUpdatedSimulation() (error) {
	jsonRes := user.simulation.getJsonRepresentation()
	message := Message{Event: simulationUpdate, Data: jsonRes}
	marshalledMessage, err := json.Marshal(message)
	if err != nil {
		panic(err)
		return err
		// TODO handle marshall err
	}
	user.write(marshalledMessage)
	return nil
}

// Below this is to handle reading and sending message from websockets
func (user *User) reader() {
	defer user.close()

	var msg Message
	for {
		if user.ws == nil {
			return
		}
		// read incoming message from socket
		if err := user.ws.ReadJSON(&msg); err != nil {
			log.Printf("socket read error: %v\n", err)
			break
		}
		if playerShowLog == true {
			log.Printf("%s -> %s\n", user.ws.RemoteAddr(), msg.Event)
		}
		fmt.Println("message recieved")
		fmt.Println(msg.Event)
		// assign message to a function handler
		if handler, found := user.group.FindHandler(msg.Event); found {
			// send msg.ID
			handler(user.ws, msg.Data)
		}
	}

	user.log("Exiting reader.")
}
func (user *User) Write(p []byte) (n int, err error) {

	if user.ws == nil {
		return 0, ErrNoWebsocket
	}

	err = user.ws.WriteMessage(websocket.TextMessage, p)

	if playerShowLog == true {
		log.Printf("%s <- %s: %v\n", user.ws.RemoteAddr(), p, err)
	}

	return len(p), err
}

func (user *User) writer() {
	var start, diff, sleep int64
	var buf []byte

	writing := true

	for writing {
		buf = make([]byte, 0)

		start = time.Now().UnixNano()

		select {
		case message := <-user.output:
			buf = append(buf, message...)
			break

		}

		if len(buf) > 0 {
			_, err := user.Write(buf)
			if err != nil {
				writing = false
			}
		}

		diff = time.Now().UnixNano() - start
		sleep = fpsn - diff

		if sleep < 0 {
			continue
		}

		time.Sleep(time.Duration(sleep) * time.Nanosecond)
	}

	user.close()
}

func (user *User) close() {
	if user == nil {
		return
	}
	if user.ws != nil {
		user.ws.Close()
		user.ws = nil
		user.group.removePlayer(user)
	}

	if user.runningSimulation {
		user.runningSimulation = false
		user.simulation.cancelSimulation <- true
		user.simulation = nil
	}

	if user.output != nil {
		close(user.output)
		user.output = nil
	}
}
