package main

import (
	"fmt"
	"github.com/go-siris/siris/core/errors"
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
}

func (loc *StatefulLocation) addNCars(numCars int, direction Direction) {
	for i := 0; i < numCars; i++ {
		id := fmt.Sprintf("car %d", i)
		loc.Cars[id] = &SmartCar{id: id, direction: direction, x: -1, y: -1}
	}
}

type GeneralLaneSimulationConfig struct {
	sizeOfLane         int
	numVerticalLanes   int
	numHorizontalLanes int
}

// GeneralLaneSimulation handles a general simulation with n horizontal lanes and n vertical lanes and intersections
type GeneralLaneSimulation struct {
	Simulation
	Locations      [][]StatefulLocation
	horizontalRoot StatefulLocation
	verticalRoot   StatefulLocation
	config         GeneralLaneSimulationConfig

	moveCarsInHorizontal  chan bool
	moveCarsOutHorizontal chan bool

	moveCarsInVertical  chan bool
	moveCarsOutVertical chan bool

	carClock chan *Car
}

func medianBasedRange(laneSize int, numLanes int) (int, int) {
	return int(laneSize/2 - numLanes/2), int(laneSize/2 + numLanes/2)
}
func (sim *GeneralLaneSimulation) horizontalIndexRange() (int, int) {
	config := sim.config
	return medianBasedRange(config.sizeOfLane, config.numHorizontalLanes)
}

func (sim *GeneralLaneSimulation) verticalIndexRange() (int, int) {
	config := sim.config
	return medianBasedRange(config.sizeOfLane, config.numVerticalLanes)
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
	locations := [sizeOfLane][sizeOfLane]StatefulLocation{}

	// Initialize horizontal locations
	horizontalStartIndex, horizontalEndIndex := simulation.horizontalIndexRange()
	for i := horizontalStartIndex; i <= horizontalEndIndex; i++ {
		for j := 0; j < sizeOfLane; j++ {
			locations[i][j].locationState = LaneLoc
		}
	}
	if numHorizontalLanes > 0 {
		horizontalRoot := StatefulLocation{Cars: make(map[string]*SmartCar, sizeOfLane)}
		horizontalRoot.addNCars(sizeOfLane, Horizontal)
		simulation.verticalRoot = horizontalRoot
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
	if numVerticalLanes > 0 {
		verticalRoot := StatefulLocation{Cars: make(map[string]*SmartCar, sizeOfLane)}
		verticalRoot.addNCars(sizeOfLane, Vertical)
		simulation.verticalRoot = verticalRoot
	}

	simulation.moveCarsInHorizontal = make(chan bool)
	simulation.moveCarsOutHorizontal = make(chan bool)

	simulation.moveCarsInVertical = make(chan bool)
	simulation.moveCarsOutVertical = make(chan bool)

	simulation.carClock = make(chan *Car)
	simulation.runningSimulation = false
	simulation.drawUpdateChan = make(chan bool)
	return &simulation, nil
}

