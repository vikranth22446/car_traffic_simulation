package main;

import (
	"fmt"
	"strings"
)

// SingleLaneSimulation handles a list of cars moving from one side to another
type SingleLaneSimulation struct {
	Simulation
	Lane            *Lane
	moveCarsInLane  chan *Lane
	carClock        chan *Car
	moveCarsEndLane chan *Lane
}

func (singleSim *SingleLaneSimulation) String() (string) {
	var b strings.Builder
	lane := singleSim.Lane
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

func initSingleLaneSimulation(sizeOfLane int) *SingleLaneSimulation {
	simulation := SingleLaneSimulation{}

	lane := Lane{
		Locations:  make([]Location, sizeOfLane),
		sizeOfLane: sizeOfLane,
	}
	simulation.Lane = &lane

	for i := 0; i < sizeOfLane; i++ {
		lane.Locations[i].Cars = make(map[string]*Car, 0)
		//carUUID := uuid.New() can use ID or car _
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


// RunSingleLaneSimulation runs the simulation such that all the cars from bin 0 move to the last bin
func RunSingleLaneSimulation(simulation *SingleLaneSimulation) {
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
			car.lanePos++
			go MoveCarInLane(car, carClock)
			drawUpdateChan <- true
			break
		}
	}
}
