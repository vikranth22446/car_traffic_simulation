package main

import (
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
	ErrNoWebsocket = errors.New(`Don't have any websocket connection.`)
)

type User struct {
	ws     *websocket.Conn
	output chan []byte
	addr   net.Addr

	group             *UserGroup
	Id                uuid.UUID
	simulationRunning bool
	simulation        *Simulation
}

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

func (self *User) log(s string) {
	log.Printf("%s: %s\n", self.addr, s)
}

func (self *User) write(data []byte) {
	if self.ws == nil {
		return
	}
	if self.output == nil {
		return
	}
	self.output <- data
}

func (self *User) Is(other *User) bool {
	return self.Id == other.Id
}

func (self *User) identify() {
	self.write(identFn(self.Id))
}

func (self *User) runSimulation() {
	if self.simulationRunning {
		return
	}
	self.simulationRunning = true
	simulation := initSimulation(10)
	self.simulation = simulation
	go RunSimulation(simulation)
	var start, diff, sleep int64

	for {
		start = time.Now().UnixNano()
		self.update()

		diff = time.Now().UnixNano() - start
		sleep = fpsn - diff

		fmt.Printf("sleep: %d, diff: %d\n", sleep, diff)

		if sleep < 0 {
			continue
		}

		time.Sleep(time.Duration(sleep) * time.Nanosecond)
	}
}

func (self *User) update() {
	//chunk := updateFn(self.Id, b)
	// TODO Update game state and send message
	//b := self.Serialize()
	//if b != nil {
	//
	//	if self.sector != nil {
	//		for p, _ := range self.sector.players {
	//			if p.ws != nil && p.output != nil {
	//				if self.isNear(p) == true {
	//					p.write(chunk)
	//				}
	//			}
	//		}
	//	}
	//}
}

// Below this is to handle reading and sending message from websockets
func (self *User) reader() {

	for {
		_, message, err := self.ws.ReadMessage()
		if err != nil {
			break
		}
		//json.Unmarshal(message, self.control)

		if playerShowLog == true {
			log.Printf("%s -> %s\n", self.ws.RemoteAddr(), message)
		}
	}

	self.log("Exiting reader.")
	self.close()
}
func (self *User) Write(p []byte) (n int, err error) {

	if self.ws == nil {
		return 0, ErrNoWebsocket
	}

	err = self.ws.WriteMessage(websocket.TextMessage, p)

	if playerShowLog == true {
		log.Printf("%s <- %s: %v\n", self.ws.RemoteAddr(), p, err)
	}

	return len(p), err
}

func (self *User) writer() {
	var start, diff, sleep int64
	var buf []byte

	writing := true

	for writing {
		buf = make([]byte, 0, 1024*10)

		start = time.Now().UnixNano()

		loop := true
		for loop {
			select {
			case message := <-self.output:
				buf = append(buf, message...)
				buf = append(buf, '\n')
			default:
				loop = false
			}
		}

		if len(buf) > 0 {
			_, err := self.Write(buf)
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

	self.close()
}

func (self *User) close() {
	if self.ws != nil {
		self.log("Closing websocket.")
		self.ws.Close()
		self.ws = nil
		self.log("Websocket closed.")
	}

	if self.output != nil {
		self.log("Closing channel.")
		close(self.output)
		self.output = nil
		self.log("Channel closed.")
	}
}
