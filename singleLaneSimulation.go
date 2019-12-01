package main;

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type Car struct {
	id      string
	speed   float64
	lanePos int
	lane    *Lane
}

func (car *Car) String() string {
	return car.id
}

type SimulationConfig struct {
	sizeOfLane int
}

type Simulation struct {
	drawUpdateChan    chan bool
	runningSimulation bool
	config            SimulationConfig
}

type SingleLaneSimulation struct {
	Simulation
	Lane            *Lane
	moveCarsInLane  chan *Lane
	carClock        chan *Car
	moveCarsEndLane chan *Lane
}

func RightPad2Len(s string, padStr string, overallLen int) string {
	var padCountInt = 1 + ((overallLen - len(padStr)) / len(padStr))
	var retStr = s + strings.Repeat(padStr, padCountInt)
	return retStr[:overallLen]
}
func (sim *SingleLaneSimulation) String() (string) {
	var b strings.Builder
	lane := sim.Lane
	fmt.Fprintf(&b, " inBin ")
	for i := 1; i < len(lane.Locations)-1; i++ {
		loc := lane.Locations[i]
		if len(loc.Cars) == 0 {
			fmt.Fprintf(&b, RightPad2Len("", "_", 8)+" ")
		} else {
			car := getCarFromLocation(&loc, false)
			fmt.Fprintf(&b, RightPad2Len(car.id, " ", 8)+" ")
		}
	}
	fmt.Fprintf(&b, " outBin ")
	for car := range lane.Locations[lane.sizeOfLane-1].Cars {
		fmt.Fprintf(&b, car+" ")
	}

	s := b.String() // no copying
	return s
}

type Lane struct {
	Locations  []Location
	sizeOfLane int
}

type Location struct {
	Cars map[string]*Car // Allows for easy removal of the car
}

func UniformRand() float64 {
	rnd := rand.Float64()
	max := 1.0
	min := 0.0

	return rnd*(max-min) + min
}

func moveCarsThroughBins(lane *Lane, movementChan chan *Lane, start bool) {
	var location Location

	if start {
		location = (lane.Locations)[0]
	} else {
		location = lane.Locations[len(lane.Locations)-2]
	}

	for {
		movementTime := rand.ExpFloat64()
		select {
		case <-time.After(time.Duration(movementTime) * time.Second):
			movementChan <- lane
			break
		default:
			if start && len(location.Cars) == 0 {
				return
			}
		}
	}
}

func MoveCarInLane(car *Car, movementChan chan *Car) {
	p := 0.5
	movementTime := rand.ExpFloat64() / car.speed
	select {
	case <-time.After(time.Duration(movementTime) * time.Second):
		if UniformRand() < p {
			movementChan <- car
			return
		} else {
			go MoveCarInLane(car, movementChan)
		}
	}

}
func getCarFromLocation(location *Location, del bool) (*Car) {
	var currCar *Car
	for k, v := range location.Cars {
		if del {
			delete(location.Cars, k)
		}
		currCar = v
		break
	}
	return currCar
}

func initSimulation(sizeOfLane int) *SingleLaneSimulation {
	simulation := SingleLaneSimulation{}

	lane := Lane{
		Locations:  make([]Location, sizeOfLane),
		sizeOfLane: sizeOfLane,
	}
	simulation.Lane = &lane

	for i := 0; i < sizeOfLane; i++ {
		lane.Locations[i].Cars = make(map[string]*Car, 0)
		//carUUID := uuid.New() can use id or car _
		id := fmt.Sprintf("car %d", i)
		lane.Locations[0].Cars[id] = &Car{id: id, lane: &lane, lanePos: 0}
	}

	simulation.moveCarsInLane = make(chan *Lane)
	simulation.carClock = make(chan *Car)
	simulation.moveCarsEndLane = make(chan *Lane)
	simulation.runningSimulation = false
	simulation.drawUpdateChan = make(chan bool)
	return &simulation
}

func (simulation *SingleLaneSimulation) close() {
	simulation.runningSimulation = false
}

func RunSimulation(simulation *SingleLaneSimulation) {
	defer simulation.close()
	lane := simulation.Lane
	moveCarsInLane := simulation.moveCarsInLane
	moveCarsEndLane := simulation.moveCarsEndLane
	carClock := simulation.carClock
	drawUpdateChan := simulation.drawUpdateChan

	go moveCarsThroughBins(lane, moveCarsInLane, true)
	go moveCarsThroughBins(lane, moveCarsEndLane, false)
	//go MoveCarInLane(Lane, moveCarsEndLane)

	for {
		if len(lane.Locations[lane.sizeOfLane-1].Cars) == lane.sizeOfLane {
			return
		}

		select {
		case inBin := <-moveCarsInLane:
			var firstLoc = inBin.Locations[0]
			var secondLoc = inBin.Locations[1]
			if len(firstLoc.Cars) == 0 {
				break
			}
			// next spot isn't free
			if len(secondLoc.Cars) != 0 {
				break
			}
			currCar := getCarFromLocation(&firstLoc, true)
			currCar.lanePos = 1
			secondLoc.Cars[currCar.id] = currCar
			go MoveCarInLane(currCar, carClock)
			drawUpdateChan <- true
			break
		case outBin := <-moveCarsEndLane:
			var secondToLastLoc = outBin.Locations[len(outBin.Locations)-2]
			var lastLoc = outBin.Locations[len(outBin.Locations)-1]
			if len(secondToLastLoc.Cars) == 0 {
				break
			}
			currCar := getCarFromLocation(&secondToLastLoc, true)
			lastLoc.Cars[currCar.id] = currCar
			drawUpdateChan <- true
			break
		case car := <-carClock:
			if car.lanePos == car.lane.sizeOfLane-2 {
				break // last position do nothing
			}
			var currLoc = car.lane.Locations[car.lanePos]
			var nextLoc = car.lane.Locations[car.lanePos+1]
			if len(nextLoc.Cars) != 0 {
				go MoveCarInLane(car, carClock) // If next position blocked, attempt to move again on a exponential clock
				break
			}

			nextLoc.Cars[car.id] = car
			delete(currLoc.Cars, car.id) // remove the car from the current Lane
			car.lanePos += 1
			go MoveCarInLane(car, carClock)
			drawUpdateChan <- true
			break
		}
	}
}
