package main;

import (
	"github.com/google/uuid"
	"math/rand"
	"time"
)

//type State int
//
//const (
//	Goal State = iota
//	East
//	South
//	West
//)

type Car struct {
	id      uuid.UUID
	speed   float64
	lanePos int
	lane    *Lane
}

type Lane struct {
	Locations  []Location
	sizeOfLane int
}

type Location struct {
	Cars map[uuid.UUID]*Car // Allows for easy removal of the car
}

func UniformRand() float64 {
	rnd := rand.Float64()
	max := 1.0
	min := 0.0

	return rnd*(max-min) + min
}

func moveCarsThroughBins(lane *Lane, movementChan chan Lane, start bool) {
	var location Location

	if start {
		location = (lane.Locations)[0]
	} else {
		location = lane.Locations[len(lane.Locations)-2]
	}

	for {
		movementTime := rand.ExpFloat64() / 10
		select {
		case <-time.After(time.Duration(movementTime) * time.Second):
			movementChan <- *lane
			break
		default:
			if start && len(location.Cars) == 0 {
				return
			}
		}
	}
}

func MoveCarInLane(car *Car, movementChan chan Car) {
	p := 0.5
	movementTime := rand.ExpFloat64() / car.speed
	select {
	case <-time.After(time.Duration(movementTime) * time.Second):
		if UniformRand() < p {
			movementChan <- *car
		}
		return
	}
}
func getCarFromLocation(location *Location) (*Car) {
	var currCar *Car
	for k, v := range location.Cars {
		delete(location.Cars, k)
		currCar = v
		break
	}
	return currCar
}

func RunSimulation() {
	sizeOfLane := 10
	//numCarsLane1 := 10
	//cars := make([]Car, numCarsLane1)
	//lane1[0].Cars = cars
	//carPool := make([]Car, 0)

	lane := Lane{
		Locations: make([]Location, sizeOfLane),
	}

	for i := 0; i <= sizeOfLane; i++ {
		lane.Locations[i].Cars = make(map[uuid.UUID]*Car, 0)
		carUUID := uuid.New()
		lane.Locations[0].Cars[carUUID] = &Car{id: carUUID, lane: &lane, lanePos: 0}
	}

	moveCarsInLane := make(chan Lane)
	carClock := make(chan Car)
	moveCarsEndLane := make(chan Lane)

	go moveCarsThroughBins(&lane, moveCarsInLane, true)
	go moveCarsThroughBins(&lane, moveCarsEndLane, false)
	//go MoveCarInLane(lane, moveCarsEndLane)

	for {
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
			currCar := getCarFromLocation(&firstLoc)
			currCar.lanePos = 1
			secondLoc.Cars[currCar.id] = currCar
			go MoveCarInLane(currCar, carClock)
			break
		case car := <-carClock:
			if car.lanePos == car.lane.sizeOfLane-1 {
				break // last position do nothing
			}
			var nextLoc = car.lane.Locations[car.lanePos+1]
			if len(nextLoc.Cars) != 0 {
				break
			}
			currCar := getCarFromLocation(&nextLoc)
			currCar.lanePos += 1
			go MoveCarInLane(currCar, carClock)
			break
		case outBin := <-moveCarsEndLane:
			var secondToLastLoc = outBin.Locations[len(outBin.Locations)-2]
			var lastLoc = outBin.Locations[len(outBin.Locations)-1]
			if len(secondToLastLoc.Cars) == 0 {
				break
			}
			currCar := getCarFromLocation(&secondToLastLoc)
			lastLoc.Cars[currCar.id] = currCar
			break
		}

	}

	//lanes := make([][]Car, 5)
	//for i := range a {
	//	a[i] = make([]Car, 5)
	//}
	//fmt.Printf("Variables %v", numCarsLeft, numHorizontalLanes)
}
