package main;

import (
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
	id      int
	speed   float64
	lanePos int
	lane    *Lane
}

type Lane struct {
	Locations  []Location
	sizeOfLane int
}

type Location struct {
	Cars map[int]Car // Allows for easy removal of the car
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
		location = lane.Locations[len(lane.Locations)-1]
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
	for {
		movementTime := rand.ExpFloat64() / car.speed
		select {
		case <-time.After(time.Duration(movementTime) * time.Second):
			if UniformRand() < p {
				movementChan <- *car
			}
			break
		}
	}
}

func RunSimulation() {
	sizeOfLane := 10
	//numCarsLane1 := 10
	lane1 := make([]Location, sizeOfLane)
	//cars := make([]Car, numCarsLane1)
	//lane1[0].Cars = cars

	lane := Lane{
		Locations: lane1,
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
			// TODO randomly select a car and move it to the next position
			println("car is ready to move out of ", inBin)
			// Item from lane x attempts to move in
			break
		case car := <-carClock:
			// TODO check if car in the next position or not
			// TODO call the function again for the car after increment pos if not end
			// TODO if last position ignore
			println("car is ready to move to the next pos", car)
			break
		case outBin := <-moveCarsEndLane:
			// TODO Move car from second to last to last lane
			println("car is ready to move out of ", outBin)
			break
		}

	}

	//lanes := make([][]Car, 5)
	//for i := range a {
	//	a[i] = make([]Car, 5)
	//}
	//fmt.Printf("Variables %v", numCarsLeft, numHorizontalLanes)
}
