package main

import (
	"fmt"
	"github.com/go-siris/siris/core/errors"
	"math/rand"
	"strings"
	"time"
)

type Direction int

const (
	Horizontal Direction = iota
	Vertical
)

type SmartCar struct {
	id        string
	speed     float64
	x         int
	y         int
	direction Direction
}

type LocationState int

const (
	Empty LocationState = iota
	Intersection
	LaneLoc
	Parking
)

// Location is one spot on a lane
type StatefulLocation struct {
	Cars          map[string]*SmartCar // Allows for easy removal of the car
	locationState LocationState
	x             int
	y             int
}

func (loc *StatefulLocation) addNCars(numCars int, direction Direction) {
	for i := 0; i < numCars; i++ {
		var id string
		if direction == Horizontal {
			id = fmt.Sprintf("hcar %d", i)
		} else {
			id = fmt.Sprintf("vcar %d", i)
		}
		loc.Cars[id] = &SmartCar{id: id, direction: direction, x: -1, y: -1, speed: 3}
	}
}

func (loc *StatefulLocation) isEmpty() bool {
	return len(loc.Cars) == 0
}

func (loc *StatefulLocation) getCar(del bool) (*SmartCar) {
	var currCar *SmartCar
	for k, v := range loc.Cars {
		if del {
			delete(loc.Cars, k)
		}
		currCar = v
		break
	}
	return currCar
}

func (loc *StatefulLocation) addCar(car *SmartCar) {
	car.x = loc.x
	car.y = loc.y
	loc.Cars[car.id] = car
}

type GeneralLaneSimulationConfig struct {
	sizeOfLane         int
	numVerticalLanes   int
	numHorizontalLanes int
}

// GeneralLaneSimulation handles a general simulation with n horizontal lanes and n vertical lanes and intersections
type GeneralLaneSimulation struct {
	Simulation
	Locations        [][]*StatefulLocation
	InHorizontalRoot *StatefulLocation
	InVerticalRoot   *StatefulLocation

	OutHorizontalRoot *StatefulLocation
	OutVerticalRoot   *StatefulLocation

	config GeneralLaneSimulationConfig

	moveCarsIn  chan Direction
	moveCarsOut chan Direction

	carClock chan *SmartCar
}

func (sim *GeneralLaneSimulation) String() (string) {
	var b strings.Builder
	fmt.Fprintf(&b, RightPad2Len("", "-", 20)+" \n")

	fmt.Fprintf(&b, " horizontalInBin ")
	for car := range sim.InHorizontalRoot.Cars {
		fmt.Fprintf(&b, car+" ")
	}
	fmt.Fprintf(&b, " horizontalOutBin ")
	for car := range sim.OutHorizontalRoot.Cars {
		fmt.Fprintf(&b, car+" ")
	}
	fmt.Fprintf(&b, "\n")
	fmt.Fprintf(&b, " verticalInBin ")
	for car := range sim.InVerticalRoot.Cars {
		fmt.Fprintf(&b, car+" ")
	}
	fmt.Fprintf(&b, " verticalOutBin ")
	for car := range sim.OutVerticalRoot.Cars {
		fmt.Fprintf(&b, car+" ")
	}

	fmt.Fprintf(&b, "\n")
	for i := 0; i < sim.config.sizeOfLane; i++ {
		for j := 0; j < sim.config.sizeOfLane; j++ {
			loc := sim.Locations[i][j]
			if len(loc.Cars) == 0 {
				fmt.Fprintf(&b, RightPad2Len("", "_", 8)+" ")
			} else {
				car := loc.getCar(false)
				fmt.Fprintf(&b, RightPad2Len(car.id, " ", 8)+" ")
			}
		}
		fmt.Fprintf(&b, "\n")
	}

	s := b.String() // no copying
	return s
}

func medianBasedRange(laneSize int, numLanes int) (int, int) {
	return int(laneSize/2 - (numLanes-1)/2), int(laneSize/2 + (numLanes-1)/2)
}

func (sim *GeneralLaneSimulation) horizontalIndexRange() (int, int) {
	config := sim.config
	return medianBasedRange(config.sizeOfLane, config.numHorizontalLanes)
}

func (sim *GeneralLaneSimulation) verticalIndexRange() (int, int) {
	config := sim.config
	return medianBasedRange(config.sizeOfLane, config.numVerticalLanes)
}

func (sim *GeneralLaneSimulation) allCarsMovedIn() (bool) {
	if sim.config.numHorizontalLanes > 0 && len(sim.OutHorizontalRoot.Cars) != sim.config.sizeOfLane {
		return false
	}
	if sim.config.numVerticalLanes > 0 && len(sim.OutVerticalRoot.Cars) != sim.config.sizeOfLane {
		return false
	}
	return true
}

func initMultiLaneSimulation(config GeneralLaneSimulationConfig) (*GeneralLaneSimulation, error) {
	simulation := GeneralLaneSimulation{config: config}
	sizeOfLane := simulation.config.sizeOfLane
	numVerticalLanes := simulation.config.numVerticalLanes
	numHorizontalLanes := simulation.config.numHorizontalLanes

	if sizeOfLane-2 <= 0 {
		return nil, errors.New("The lane must be 2 spots")
	}
	if !(sizeOfLane > numVerticalLanes && sizeOfLane > numHorizontalLanes) {
		return nil, errors.New("The number of vertical/horizontal lanes cannot be more than size of lane")
	}
	locations := make([][] *StatefulLocation, sizeOfLane)
	for i := range locations {
		locations[i] = make([]*StatefulLocation, sizeOfLane)
		for j := 0; j < sizeOfLane; j++ {
			locations[i][j] = &StatefulLocation{locationState: Empty, Cars: make(map[string]*SmartCar)}
			locations[i][j].x = i
			locations[i][j].y = j
		}
	}

	// Initialize horizontal locations
	horizontalStartIndex, horizontalEndIndex := simulation.horizontalIndexRange()
	for i := horizontalStartIndex; i <= horizontalEndIndex; i++ {
		for j := 0; j < sizeOfLane; j++ {
			locations[i][j].locationState = LaneLoc
			locations[i][j].Cars = make(map[string]*SmartCar, 0)
		}
	}
	if numHorizontalLanes > 0 {
		horizontalRoot := StatefulLocation{Cars: make(map[string]*SmartCar, 0)}
		horizontalRoot.addNCars(sizeOfLane, Horizontal)
		simulation.InHorizontalRoot = &horizontalRoot
		simulation.OutHorizontalRoot = &StatefulLocation{Cars: make(map[string]*SmartCar, 0)}
	}

	// Initialize vertical locations
	verticalStartIndex, verticalEndIndex := simulation.verticalIndexRange()
	for i := 0; i < sizeOfLane; i++ {
		for j := verticalStartIndex; j <= verticalEndIndex; j++ {
			if locations[i][j].locationState == LaneLoc {
				locations[i][j].locationState = Intersection
				// If it is already on the horizontal path then send in update
			} else {
				locations[i][j].locationState = LaneLoc
			}
		}
	}
	simulation.Locations = locations

	if numVerticalLanes > 0 {
		verticalRoot := StatefulLocation{Cars: make(map[string]*SmartCar, 0)}
		verticalRoot.addNCars(sizeOfLane, Vertical)
		simulation.InVerticalRoot = &verticalRoot
		simulation.OutVerticalRoot = &StatefulLocation{Cars: make(map[string]*SmartCar, 0)}

	}

	simulation.moveCarsIn = make(chan Direction)
	simulation.moveCarsOut = make(chan Direction)

	simulation.carClock = make(chan *SmartCar)
	simulation.runningSimulation = false
	simulation.drawUpdateChan = make(chan bool)
	return &simulation, nil
}
func RandomlyPickLocation(lanes []*StatefulLocation) *StatefulLocation {
	return lanes[rand.Intn(len(lanes))]
}

// RunSingleLaneSimulation runs the simulation such that all the cars from bin 0 move to the last bin
func RunGeneralSimulation(simulation *GeneralLaneSimulation) {
	defer simulation.close()

	moveCarsIn := simulation.moveCarsIn
	moveCarsOut := simulation.moveCarsOut

	carClock := simulation.carClock
	drawUpdateChan := simulation.drawUpdateChan

	go moveCarsThroughBinsDirection(moveCarsIn, Horizontal, true, simulation.InHorizontalRoot)
	go moveCarsThroughBinsDirection(moveCarsOut, Horizontal, false, simulation.OutHorizontalRoot)

	go moveCarsThroughBinsDirection(moveCarsIn, Vertical, true, simulation.InVerticalRoot)
	go moveCarsThroughBinsDirection(moveCarsOut, Vertical, false, simulation.OutHorizontalRoot)

	for {
		if len(simulation.OutHorizontalRoot.Cars) == simulation.config.sizeOfLane &&
			len(simulation.OutVerticalRoot.Cars) == simulation.config.sizeOfLane {
			return
		}
		select {
		case carIn := <-moveCarsIn:
			openLanes := make([]*StatefulLocation, 0)
			var root *StatefulLocation
			if carIn == Horizontal {
				low, high := simulation.horizontalIndexRange()
				for i := low; i <= high; i++ {
					loc := simulation.Locations[i][0]
					if loc.isEmpty() {
						openLanes = append(openLanes, loc)
					}
				}
				root = simulation.InHorizontalRoot
			} else if carIn == Vertical {
				low, high := simulation.verticalIndexRange()
				for i := low; i <= high; i++ {
					loc := simulation.Locations[0][i]
					if loc.isEmpty() {
						openLanes = append(openLanes, loc)
					}
				}
				root = simulation.InVerticalRoot
			}

			if len(openLanes) == 0 {
				break
			}
			chosenLoc := RandomlyPickLocation(openLanes)
			currCar := root.getCar(true)
			if currCar == nil {
				break
			}
			chosenLoc.addCar(currCar)
			go MoveSmartCarInLane(currCar, carClock)
			drawUpdateChan <- true
			break
		case carOut := <-moveCarsOut:
			openLanes := make([]*StatefulLocation, 0)
			var root *StatefulLocation
			lastIndex := simulation.config.sizeOfLane - 1
			if carOut == Horizontal {
				low, high := simulation.horizontalIndexRange()

				for i := low; i <= high; i++ {
					loc := simulation.Locations[i][lastIndex]
					if !loc.isEmpty() {
						openLanes = append(openLanes, loc)
					}
				}

				root = simulation.OutHorizontalRoot
			} else if carOut == Vertical {
				low, high := simulation.verticalIndexRange()
				for i := low; i <= high; i++ {
					loc := simulation.Locations[lastIndex][i]
					if !loc.isEmpty() {
						openLanes = append(openLanes, loc)
					}
				}
				root = simulation.OutVerticalRoot
			}

			if len(openLanes) == 0 {
				break
			}

			chosenLoc := RandomlyPickLocation(openLanes)
			currCar := chosenLoc.getCar(true)
			if currCar == nil {
				break
			}
			root.addCar(currCar)
			drawUpdateChan <- true

			break
		case car := <-carClock:
			currLoc := simulation.Locations[car.x][car.y]
			var nextLoc *StatefulLocation

			if car.direction == Horizontal {
				if car.y+1 == simulation.config.sizeOfLane {
					break
				}
				nextLoc = simulation.Locations[car.x][car.y+1]
			} else if car.direction == Vertical {
				if car.x+1 == simulation.config.sizeOfLane {
					break
				}
				nextLoc = simulation.Locations[car.x+1][car.y]
			}

			if len(nextLoc.Cars) != 0 {
				go MoveSmartCarInLane(car, carClock) // If next position blocked, attempt to move again on a exponential clock
				break
			}
			currCar := currLoc.getCar(true)
			if currCar == nil {
				break
			}
			nextLoc.addCar(currCar)
			go MoveSmartCarInLane(car, carClock) // If next position blocked, attempt to move again on a exponential clock

			drawUpdateChan <- true
			break
		}
	}
}


// MoveCarInLane moves the car through a lane using an exponential clock and probability of movement
func MoveSmartCarInLane(car *SmartCar, movementChan chan *SmartCar) {
	p := 0.5
	movementTime := rand.ExpFloat64() / car.speed
	select {
	case <-time.After(time.Duration(movementTime) * time.Second):
		if UniformRand() < p {
			movementChan <- car
			return
		}
		go MoveSmartCarInLane(car, movementChan)
	}
}

func moveCarsThroughBinsDirection(
	movementChan chan Direction,
	direction Direction,
	start bool,
	location *StatefulLocation) {
	for {
		movementTime := rand.ExpFloat64() / 2
		select {
		case <-time.After(time.Duration(movementTime) * time.Second):
			movementChan <- direction
			break
		default:
			if start && len(location.Cars) == 0 {
				return
			}
		}
	}
}
