package main

import (
	"math/rand"
	"time"
)

// Car the main object that is moved through a lane
type Car struct {
	id      string
	speed   float64
	lanePos int
	lane    *Lane
}

func (car *Car) String() string {
	return car.id
}

// SimulationConfig is the config that is used in a simulation
type SimulationConfig struct {
	sizeOfLane int
}

// Simulation handles the channel to get updates from and the configuration
type Simulation struct {
	drawUpdateChan    chan bool
	runningSimulation bool
	config            SimulationConfig
}

func (singleSim *Simulation) close() {
	singleSim.runningSimulation = false
}

// Lane holds a list of locations
type Lane struct {
	Locations  []Location
	sizeOfLane int
}

// Location is one spot on a lane
type Location struct {
	Cars map[string]*Car // Allows for easy removal of the car
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

// MoveCarInLane moves the car through a lane using an exponential clock and probability of movement
func MoveCarInLane(car *Car, movementChan chan *Car) {
	p := 0.5
	movementTime := rand.ExpFloat64() / car.speed
	select {
	case <-time.After(time.Duration(movementTime) * time.Second):
		if UniformRand() < p {
			movementChan <- car
			return
		}
		go MoveCarInLane(car, movementChan)
	}
}

//
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
