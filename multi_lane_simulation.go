package main

import (
	"fmt"
	"github.com/go-siris/siris/core/errors"
	"gonum.org/v1/gonum/stat/distuv"
	"gonum.org/v1/gonum/stat/sampleuv"
	"math/rand"
	"strings"
	"sync"
	"time"
)

type Direction int

const (
	Horizontal Direction = iota
	Vertical
)

type AccidentResolution int

const (
	Unresolved AccidentResolution = iota
	Resolved
	ToBeDeleted
)

type Accident struct {
	loc               *StatefulLocation
	resolution        AccidentResolution
	prevLocationState LocationState
	probRestart       float64
	removalRate       float64
}

type Parking struct {
	prevLoc         *StatefulLocation
	car             *SmartCar
	parkingTimeRate float64
	parkingLoc      *StatefulLocation
}

type SmartCarState int

const (
	Working SmartCarState = iota
	Deleted
)

type SmartCar struct {
	ID           string
	Speed        float64
	X            int
	Y            int
	Direction    Direction
	probMovement float64
	carState     SmartCarState
	slowingDown  bool
	smartCarLock sync.Mutex
}

func (car *SmartCar) isSlowingDown() bool {
	car.smartCarLock.Lock()
	defer car.smartCarLock.Unlock()
	return car.slowingDown
}

func (car *SmartCar) getSpeed() float64 {
	car.smartCarLock.Lock()
	defer car.smartCarLock.Unlock()
	return car.Speed
}

func (car *SmartCar) setSlowingDown(item bool) {
	car.smartCarLock.Lock()
	defer car.smartCarLock.Unlock()
	car.slowingDown = item
}

func (car *SmartCar) setSpeed(res float64) {
	car.smartCarLock.Lock()
	defer car.smartCarLock.Unlock()
	car.Speed = res
}

type LocationState int

const (
	Empty                 LocationState = 0
	Intersection          LocationState = 1
	LaneLoc               LocationState = 2
	ParkingLoc            LocationState = 3
	AccidentLocationState LocationState = 4
	CrossWalk             LocationState = 5
)

// Location is one spot on a lane
type StatefulLocation struct {
	Cars          map[string]*SmartCar // Allows for easy removal of the car
	LocationState LocationState
	X             int
	Y             int
	locationLock  sync.Mutex
}

func (loc *StatefulLocation) setLocationState(state LocationState) {
	loc.locationLock.Lock()
	defer loc.locationLock.Unlock()
	loc.LocationState = state
}

func (loc *StatefulLocation) getLocationState() LocationState {
	loc.locationLock.Lock()
	defer loc.locationLock.Unlock()
	return loc.LocationState
}

func getNewCarSpeed(speedType CarDistributionType, carSpeedEndRange float64) (float64, float64) {
	var speed float64
	var prob float64
	if speedType == constantDistribution {
		speed = 1.0
		prob = 1.0
	} else if speedType == normalDistribution {
		var UnitNormal = distuv.Normal{Mu: 0, Sigma: 1}
		speed = UnitNormal.Rand()
		prob = UnitNormal.Prob(speed)
	} else if speedType == exponentialDistribution {
		var exponential = distuv.Exponential{Rate: 1}
		speed = exponential.Rand()
		prob = exponential.Prob(speed)
	} else if speedType == poissonDistribution {
		var poisson = distuv.Poisson{Lambda: 1}
		speed = poisson.Rand()
		prob = poisson.Prob(speed)
	} else if speedType == uniformDistribution {
		if carSpeedEndRange == 0 {
			carSpeedEndRange = 1
		}
		speed = UniformRandMinMax(0, carSpeedEndRange)
		prob = 1 / carSpeedEndRange
	}
	return speed, prob
}

type SlowCar struct {
	car          *SmartCar
	oldSpeed     float64
	slowDownRate float64
}

func (loc *StatefulLocation) canMoveToParking() bool {
	state := loc.getLocationState()
	return state == LaneLoc || state == CrossWalk || state == AccidentLocationState
}

func (loc *StatefulLocation) addNCars(numCars int,
	direction Direction,
	probMovement float64,
	speedType CarDistributionType,
	carSpeedEndRange float64) {
	for i := 0; i < numCars; i++ {
		var id string
		if direction == Horizontal {
			id = fmt.Sprintf("hcar %d", i)
		} else {
			id = fmt.Sprintf("vcar %d", i)
		}
		speed, _ := getNewCarSpeed(speedType, carSpeedEndRange)
		loc.Cars[id] = &SmartCar{
			ID:        id,
			Direction: direction,
			X:         -1, Y: -1,
			Speed:        speed,
			probMovement: probMovement,
			carState:     Working,
			smartCarLock: sync.Mutex{}}
	}
}
func (loc *StatefulLocation) noCars() bool {
	loc.locationLock.Lock()
	defer loc.locationLock.Unlock()

	return len(loc.Cars) == 0
}
func (loc *StatefulLocation) isEmpty() bool {
	state := loc.getLocationState()
	return loc.noCars() &&
		(state == LaneLoc ||
			state == Intersection ||
			state == CrossWalk)
}

func (loc *StatefulLocation) removeCar(car *SmartCar) {
	loc.locationLock.Lock()
	defer loc.locationLock.Unlock()
	delete(loc.Cars, car.ID)
}

func (loc *StatefulLocation) getCar(del bool) (*SmartCar) {
	loc.locationLock.Lock()
	defer loc.locationLock.Unlock()

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
	loc.locationLock.Lock()
	car.smartCarLock.Lock()

	defer car.smartCarLock.Unlock()
	defer loc.locationLock.Unlock()

	car.X = loc.X
	car.Y = loc.Y
	loc.Cars[car.ID] = car

}

type CarDistributionType int

const (
	exponentialDistribution CarDistributionType = iota
	normalDistribution
	poissonDistribution
	constantDistribution
	uniformDistribution
)

func convertToCarDistributionType(item int) CarDistributionType {
	switch item {
	case 0:
		return exponentialDistribution
	case 1:
		return normalDistribution
	case 2:
		return poissonDistribution
	case 3:
		return constantDistribution
	case 4:
		return uniformDistribution
	}
	return uniformDistribution
}

type LaneChoice int

const (
	trafficBasedChoice LaneChoice = iota
	uniformLaneChoice
)

func convertIntLaneChoice(item int) LaneChoice {
	switch item {
	case 0:
		return uniformLaneChoice
	case 1:
		return trafficBasedChoice
	}
	return uniformLaneChoice
}

type GeneralLaneSimulationConfig struct {
	sizeOfLane         int
	numVerticalLanes   int
	numHorizontalLanes int

	inAlpha      float64
	outBeta      float64
	carMovementP float64

	// num cars to through in each bin
	numHorizontalCars int
	numVerticalCars   int

	// in lane movement type
	inLaneChoice  LaneChoice
	outLaneChoice LaneChoice

	// multiple lanes
	probSwitchingLanes float64
	laneSwitchChoice   LaneChoice

	// Handles accidents
	accidentProb float64

	carRemovalRate float64 // exponential time
	carRestartProb float64 // car restart rate

	// Handles different car rates
	carClock                float64 // for uniform, defaults to start of range
	carSpeedUniformEndRange float64
	CarDistributionType     CarDistributionType
	reSampleSpeedEveryClk   bool

	// handles cars going to fast
	probPolicePullOverProb float64
	speedBasedPullOver     bool

	// parking
	parkingEnabled  bool
	distractionRate float64 // poisson to get into parking
	parkingTimeRate float64 // how long should car stay in parking
	crossWalkCutoff int     /// numbers of cars in parking

	// crosswalk
	crossWalkEnabled      bool
	crossWalkSlowDownRate float64

	pedestrianDeathAccidentProb float64 //

	// intersection
	probEnteringIntersection float64
	intersectionAccidentProb float64 // if unspecified the same as regular accident probability
	accidentScaling          bool    // retry for accident based on the number of cars there
	slowDownSpeed            float64
	// scales poisson rate by certain amount
}

func DefaultGeneralLaneConfig() *GeneralLaneSimulationConfig {
	config := GeneralLaneSimulationConfig{}
	config.sizeOfLane = 10
	config.numHorizontalLanes = 1
	config.numVerticalLanes = 1
	config.inAlpha = 1
	config.outBeta = 1
	config.inLaneChoice = uniformLaneChoice
	config.outLaneChoice = uniformLaneChoice

	config.numHorizontalCars = 10
	config.numVerticalCars = 10

	config.carMovementP = 0.5

	config.probSwitchingLanes = 0

	config.accidentProb = 0

	config.carClock = 1
	config.CarDistributionType = constantDistribution
	config.reSampleSpeedEveryClk = false

	config.parkingEnabled = false
	config.probEnteringIntersection = 1
	config.parkingTimeRate = 1
	config.accidentScaling = false
	config.crossWalkCutoff = 2

	config.intersectionAccidentProb = 0
	return &config
}

// GeneralLaneSimulation handles a general simulation with n horizontal lanes and n vertical lanes and intersections
type GeneralLaneSimulation struct {
	Simulation
	Locations        [][]*StatefulLocation
	InHorizontalRoot *StatefulLocation
	InVerticalRoot   *StatefulLocation

	OutHorizontalRoot *StatefulLocation
	OutVerticalRoot   *StatefulLocation

	config *GeneralLaneSimulationConfig

	moveCarsIn  chan Direction
	moveCarsOut chan Direction

	carClock       chan *SmartCar
	crossWalkClock chan *SlowCar

	accidentChan chan *Accident

	parkingReturn chan *Parking

	runningSimulationLock sync.Mutex
	numAccidents          int
}

func (sim *GeneralLaneSimulation) isRunningSimulation() bool {
	sim.runningSimulationLock.Lock()
	defer sim.runningSimulationLock.Unlock()
	return sim.runningSimulation
}

func (sim *GeneralLaneSimulation) setRunningSimulation(res bool) {
	sim.runningSimulationLock.Lock()
	defer sim.runningSimulationLock.Unlock()
	sim.runningSimulation = res
}

type JsonGeneralLocation struct {
	Cars          map[string]SmartCar `json:"cars"` // Allows for easy removal of the car
	LocationState int                 `json:"state"`
}

type JsonGeneralLaneSimulation struct {
	Locations [][]JsonGeneralLocation `json:"locations"`
}

func (sim *GeneralLaneSimulation) getJsonRepresentation() JsonGeneralLaneSimulation {
	jsonGen := JsonGeneralLaneSimulation{Locations: make([][]JsonGeneralLocation, sim.config.sizeOfLane)}
	for i := 0; i < sim.config.sizeOfLane; i++ {
		jsonGen.Locations[i] = make([]JsonGeneralLocation, sim.config.sizeOfLane)
		for j := 0; j < sim.config.sizeOfLane; j++ {
			jsonGen.Locations[i][j] = JsonGeneralLocation{}
			jsonGen.Locations[i][j].Cars = make(map[string]SmartCar, 0)
			jsonGen.Locations[i][j].LocationState = int(sim.Locations[i][j].getLocationState())
			loc := sim.Locations[i][j]

			loc.locationLock.Lock()
			cars := loc.Cars
			for k, v := range cars {
				jsonGen.Locations[i][j].Cars[k] = SmartCar{ID: v.ID}
			}
			loc.locationLock.Unlock()

		}
	}

	return jsonGen
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
			loc.locationLock.Lock()
			if len(loc.Cars) == 0 {
				fmt.Fprintf(&b, RightPad2Len("", "_", 8)+" ")
			} else {
				car := loc.getCar(false)
				fmt.Fprintf(&b, RightPad2Len(car.ID, " ", 8)+" ")
			}
			loc.locationLock.Unlock()
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
	if config.numHorizontalLanes%2 == 0 {
		left, right := medianBasedRange(config.sizeOfLane-1, config.numHorizontalLanes)
		return left - 1, right
	}
	return medianBasedRange(config.sizeOfLane, config.numHorizontalLanes)
}

func (sim *GeneralLaneSimulation) verticalIndexRange() (int, int) {
	config := sim.config
	if config.numVerticalLanes%2 == 0 {
		left, right := medianBasedRange(config.sizeOfLane-1, config.numVerticalLanes)
		return left - 1, right
	}
	return medianBasedRange(config.sizeOfLane, config.numVerticalLanes)
}

func (sim *GeneralLaneSimulation) allCarsMovedIn() (bool) {
	sim.OutHorizontalRoot.locationLock.Lock()
	sim.OutVerticalRoot.locationLock.Lock()

	defer sim.OutHorizontalRoot.locationLock.Unlock()
	defer sim.OutVerticalRoot.locationLock.Unlock()

	if sim.config.numHorizontalLanes > 0 && len(sim.OutHorizontalRoot.Cars) != sim.config.sizeOfLane {
		return false
	}
	if sim.config.numVerticalLanes > 0 && len(sim.OutVerticalRoot.Cars) != sim.config.sizeOfLane {
		return false
	}
	return true
}

type CheckLocationType int

const (
	Open CheckLocationType = iota
	NotOpen
	AllLocationTypes
)

func (loc *StatefulLocation) shouldCheckForMovingIn(checkLocationType CheckLocationType) bool {
	return checkLocationType == Open && loc.isEmpty() || checkLocationType == NotOpen && !loc.isEmpty() || checkLocationType == AllLocationTypes
}

func (sim *GeneralLaneSimulation) getHorizontalLanesAtIndex(index int, checkLocationType CheckLocationType) []*StatefulLocation {
	openLanes := make([]*StatefulLocation, 0)
	low, high := sim.horizontalIndexRange()
	for i := low; i <= high; i++ {
		loc := sim.Locations[i][index]
		if loc.shouldCheckForMovingIn(checkLocationType) {
			openLanes = append(openLanes, loc)
		}
	}
	return openLanes
}

func (sim *GeneralLaneSimulation) getVerticalLanesAtIndex(index int, checkLocationType CheckLocationType) []*StatefulLocation {
	openLanes := make([]*StatefulLocation, 0)
	low, high := sim.verticalIndexRange()
	for i := low; i <= high; i++ {
		loc := sim.Locations[index][i]
		if loc.shouldCheckForMovingIn(checkLocationType) {
			openLanes = append(openLanes, loc)
		}
	}
	return openLanes
}

func initMultiLaneSimulation(config *GeneralLaneSimulationConfig) (*GeneralLaneSimulation, error) {
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
			locations[i][j] = &StatefulLocation{LocationState: Empty, Cars: make(map[string]*SmartCar), locationLock: sync.Mutex{}}
			locations[i][j].X = i
			locations[i][j].Y = j
		}
	}

	// Initialize horizontal locations
	horizontalStartIndex, horizontalEndIndex := simulation.horizontalIndexRange()
	for i := horizontalStartIndex; i <= horizontalEndIndex; i++ {
		for j := 0; j < sizeOfLane; j++ {
			locations[i][j].LocationState = LaneLoc
		}

	}

	if numHorizontalLanes > 0 {
		horizontalRoot := StatefulLocation{Cars: make(map[string]*SmartCar, 0)}
		horizontalRoot.addNCars(simulation.config.numHorizontalCars,
			Horizontal,
			simulation.config.carMovementP,
			simulation.config.CarDistributionType,
			simulation.config.carSpeedUniformEndRange)
		simulation.InHorizontalRoot = &horizontalRoot
		simulation.OutHorizontalRoot = &StatefulLocation{Cars: make(map[string]*SmartCar, 0), X: -1, Y: -1}
	}

	// Initialize vertical locations
	verticalStartIndex, verticalEndIndex := simulation.verticalIndexRange()
	for i := 0; i < sizeOfLane; i++ {
		for j := verticalStartIndex; j <= verticalEndIndex; j++ {
			if locations[i][j].LocationState == LaneLoc {
				locations[i][j].LocationState = Intersection
				// If it is already on the horizontal path then send in update
			} else {
				locations[i][j].LocationState = LaneLoc
			}
		}
	}
	simulation.Locations = locations

	if numVerticalLanes > 0 {
		verticalRoot := StatefulLocation{Cars: make(map[string]*SmartCar, 0)}
		verticalRoot.addNCars(simulation.config.numVerticalCars,
			Vertical,
			simulation.config.carMovementP,
			simulation.config.CarDistributionType,
			simulation.config.carSpeedUniformEndRange)
		simulation.InVerticalRoot = &verticalRoot
		simulation.OutVerticalRoot = &StatefulLocation{Cars: make(map[string]*SmartCar, 0), X: -1, Y: -1}
	}

	// Note: Decided to only use one position for parking
	// Initialize ParkingLoc Locations
	if simulation.config.parkingEnabled {
		for i := 0; i < sizeOfLane; i++ {
			//verticalParkingStart := verticalStartIndex - 1
			verticalParkingEnd := verticalEndIndex + 1
			//simulation.addParkingIfInBounds(verticalParkingStart, i)
			simulation.addParkingIfInBounds(verticalParkingEnd, i)

			//horizontalParkingStart := horizontalStartIndex - 1
			horizontalParkingEnd := horizontalEndIndex + 1
			//simulation.addParkingIfInBounds(i, horizontalParkingStart)
			simulation.addParkingIfInBounds(i, horizontalParkingEnd)
		}
	}

	simulation.moveCarsIn = make(chan Direction)
	simulation.moveCarsOut = make(chan Direction)

	simulation.carClock = make(chan *SmartCar)
	simulation.accidentChan = make(chan *Accident)
	simulation.parkingReturn = make(chan *Parking)
	simulation.crossWalkClock = make(chan *SlowCar)

	simulation.setRunningSimulation(false)
	simulation.drawUpdateChan = make(chan bool)

	simulation.runningSimulationLock = sync.Mutex{}

	return &simulation, nil
}
func normalize(arr []float64, totalCount float64) []float64 {
	weights := make([]float64, 0)
	for _, item := range arr {
		weights = append(weights, item/totalCount)
	}
	return weights
}
func (sim *GeneralLaneSimulation) addParkingIfInBounds(i int, j int) {
	if !isInBounds(i, sim.config.sizeOfLane) || !isInBounds(j, sim.config.sizeOfLane) {
		return
	}
	location := sim.Locations[i][j]
	state := location.getLocationState()
	if state == Intersection || state == LaneLoc {
		return
	}
	location.setLocationState(ParkingLoc)
}
func isInBounds(index int, size int) bool {
	return index >= 0 && index < size
}
func (sim *GeneralLaneSimulation) countNumCarsNearby(loc *StatefulLocation) int {
	count := 1
	if isInBounds(loc.X+1, sim.config.sizeOfLane) && !sim.Locations[loc.X+1][loc.Y].isEmpty() {
		count++
	}
	if isInBounds(loc.X-1, sim.config.sizeOfLane) && !sim.Locations[loc.X-1][loc.Y].isEmpty() {
		count++
	}
	if isInBounds(loc.Y+1, sim.config.sizeOfLane) && !sim.Locations[loc.X-1][loc.Y+1].isEmpty() {
		count++
	}
	if isInBounds(loc.Y-1, sim.config.sizeOfLane) && !sim.Locations[loc.X-1][loc.Y-1].isEmpty() {
		count++
	}
	return count
}
func (sim *GeneralLaneSimulation) selectBasedOnTraffic(locs []*StatefulLocation, direction Direction) *StatefulLocation {
	weights := make([]float64, 0)
	totalCount := 0.0
	var i int
	for _, item := range locs {
		if direction == Horizontal {
			i = item.X
		} else {
			i = item.Y
		}

		count := 0.0
		for j := 0; j < sim.config.sizeOfLane; j++ {
			if direction == Horizontal && !sim.Locations[i][j].isEmpty() ||
				direction == Vertical && !sim.Locations[j][i].isEmpty() {
				count++
			}
		}
		weights = append(weights, count)

		totalCount += count
	}
	weights = normalize(weights, totalCount)

	w := sampleuv.NewWeighted(
		weights,
		nil,
	)

	i, _ = w.Take()
	return locs[i%len(locs)]
}

func (sim *GeneralLaneSimulation) RandomlyPickLocation(lanes []*StatefulLocation, direction Direction, choice LaneChoice) *StatefulLocation {
	if choice == uniformLaneChoice {
		return lanes[rand.Intn(len(lanes))]
	}
	return sim.selectBasedOnTraffic(lanes, direction)
}

func (singleSim *GeneralLaneSimulation) close() {
	singleSim.setRunningSimulation(false)
}

// RunSingleLaneSimulation runs the simulation such that all the cars from bin 0 move to the last bin
func RunGeneralSimulation(simulation *GeneralLaneSimulation) {
	defer simulation.close()
	simulation.setRunningSimulation(true)

	moveCarsIn := simulation.moveCarsIn
	moveCarsOut := simulation.moveCarsOut

	carClock := simulation.carClock
	accidentChan := simulation.accidentChan
	parkingChan := simulation.parkingReturn
	crossWalkClock := simulation.crossWalkClock

	drawUpdateChan := simulation.drawUpdateChan

	go moveCarsThroughBinsDirection(moveCarsIn, Horizontal, true, simulation.InHorizontalRoot, simulation.config.inAlpha)
	go moveCarsThroughBinsDirection(moveCarsOut, Horizontal, false, simulation.OutHorizontalRoot, simulation.config.outBeta)

	go moveCarsThroughBinsDirection(moveCarsIn, Vertical, true, simulation.InVerticalRoot, simulation.config.inAlpha)
	go moveCarsThroughBinsDirection(moveCarsOut, Vertical, false, simulation.OutHorizontalRoot, simulation.config.outBeta)
	fmt.Println("starting simulation")
	for {
		if !simulation.isRunningSimulation() {
			return
		}

		if simulation.isCompleted() {
			return
		}

		select {
		case carInDirection := <-moveCarsIn:
			if !simulation.isRunningSimulation() {
				return
			}
			openLanes := make([]*StatefulLocation, 0)
			var root *StatefulLocation
			if carInDirection == Horizontal {
				openLanes = simulation.getHorizontalLanesAtIndex(0, Open)
				root = simulation.InHorizontalRoot
			} else if carInDirection == Vertical {
				openLanes = simulation.getVerticalLanesAtIndex(0, Open)
				root = simulation.InVerticalRoot
			}

			if len(openLanes) == 0 {
				break
			}

			chosenLoc := simulation.RandomlyPickLocation(openLanes, carInDirection, simulation.config.inLaneChoice)

			currCar := root.getCar(true) // allows for picking any car from the pool
			if currCar == nil {
				break
			}

			chosenLoc.addCar(currCar)
			//fmt.Println("placing car ", currCar.ID, "at", currCar.X, currCar.Y)
			go MoveSmartCarInLane(currCar, carClock, chosenLoc)
			drawUpdateChan <- true
			break
		case carOutDirection := <-moveCarsOut:
			if !simulation.isRunningSimulation() {
				return
			}
			openLanes := make([]*StatefulLocation, 0)
			var root *StatefulLocation
			lastIndex := simulation.config.sizeOfLane - 1
			if carOutDirection == Horizontal {
				openLanes = simulation.getHorizontalLanesAtIndex(lastIndex, NotOpen)
				root = simulation.OutHorizontalRoot
			} else if carOutDirection == Vertical {
				openLanes = simulation.getVerticalLanesAtIndex(lastIndex, NotOpen)
				root = simulation.OutVerticalRoot
			}

			if len(openLanes) == 0 {
				break
			}

			chosenLoc := simulation.RandomlyPickLocation(openLanes, carOutDirection, simulation.config.outLaneChoice)
			currCar := chosenLoc.getCar(true) // allows for removing any car from the pool

			if currCar == nil {
				break
			}

			root.addCar(currCar)
			//fmt.Println("took out car", currCar.ID, currCar.X, currCar.Y)
			drawUpdateChan <- true

			break

		case car := <-carClock:
			if !simulation.isRunningSimulation() {
				return
			}
			if car == nil {
				break
			}
			car.smartCarLock.Lock()
			x := car.X
			y := car.Y
			direction := car.Direction
			car.smartCarLock.Unlock()

			if x == -1 || y == -1 ||
				!isInBounds(x, simulation.config.sizeOfLane) ||
				!isInBounds(y, simulation.config.sizeOfLane) {
				break
			}
			currLoc := simulation.Locations[x][y]

			var nextLoc *StatefulLocation
			switchLanes := UniformRand() < simulation.config.probSwitchingLanes

			if direction == Horizontal {
				if y+1 == simulation.config.sizeOfLane {
					break
				}
				if !switchLanes {
					nextLoc = simulation.Locations[x][y+1]
				} else {
					var openLanes []*StatefulLocation
					openLanes = simulation.getHorizontalLanesAtIndex(y+1, AllLocationTypes)
					nextLoc = simulation.RandomlyPickLocation(openLanes, direction, simulation.config.laneSwitchChoice) // TODO consider whether the car can pick its own position to switch to
				}
			} else if direction == Vertical {
				if x+1 == simulation.config.sizeOfLane {
					break
				}
				if !switchLanes {
					nextLoc = simulation.Locations[x+1][y]
				} else {
					var openLanes []*StatefulLocation
					openLanes = simulation.getVerticalLanesAtIndex(x+1, AllLocationTypes)
					nextLoc = simulation.RandomlyPickLocation(openLanes, direction, simulation.config.laneSwitchChoice)
				}
			}
			if currLoc.getLocationState() == AccidentLocationState {
				break
			}
			if nextLoc.getLocationState() == AccidentLocationState {
				go MoveSmartCarInLane(car, carClock, nextLoc) // just try again later
				break
			}

			if simulation.config.parkingEnabled && currLoc.canMoveToParking() { // parking can only happen on regular lane
				poisson := distuv.Poisson{Lambda: 1}
				distractionOccurs := poisson.Prob(poisson.Rand()) < simulation.config.distractionRate
				if distractionOccurs {
					var parkingLoc *StatefulLocation
					if direction == Horizontal {
						_, bottomEnd := simulation.horizontalIndexRange()
						if isInBounds(bottomEnd+1, simulation.config.sizeOfLane) {
							parkingLoc = simulation.Locations[bottomEnd+1][y]
						}
					} else {
						_, bottomEnd := simulation.verticalIndexRange()
						if isInBounds(bottomEnd+1, simulation.config.sizeOfLane) {
							parkingLoc = simulation.Locations[x][bottomEnd+1]
						}
					}
					if parkingLoc != nil {
						currLoc.removeCar(car)
						parkingLoc.addCar(car)
						go HandleParking(&Parking{prevLoc: currLoc, car: car, parkingTimeRate: simulation.config.parkingTimeRate, parkingLoc: parkingLoc}, parkingChan)
						simulation.AddCrossWalkIfNeeded(parkingLoc, direction)
						break
					}
				}
			}

			var accidentOccurs = false
			if !nextLoc.noCars() {

				poisson := distuv.Poisson{Lambda: 1}
				accidentOccurs = poisson.Prob(poisson.Rand()) < simulation.config.accidentProb

				if nextLoc.getLocationState() == Intersection {
					accidentOccurs = poisson.Prob(poisson.Rand()) < simulation.config.intersectionAccidentProb
				}

				//fmt.Println("next Car has more than 1", accidentOccurs, poisson.Prob(poisson.Rand()))

				if simulation.config.accidentScaling {
					numCarsNearby := simulation.countNumCarsNearby(nextLoc)
					for i := 0; i < numCarsNearby; i++ {
						if accidentOccurs {
							break
						}
						accidentOccurs = poisson.Prob(poisson.Rand()) < simulation.config.accidentProb
					}
				}

				// Handling of possible accident

				if !accidentOccurs {
					go MoveSmartCarInLane(car, carClock, nextLoc)
					break
				}
				// If next position blocked, attempt to move again on a exponential clock
			}

			if !accidentOccurs && nextLoc.getLocationState() == CrossWalk {
				poisson := distuv.Poisson{Lambda: 1}
				accidentOccurs = poisson.Prob(poisson.Rand()) < simulation.config.pedestrianDeathAccidentProb
			}

			if accidentOccurs {
				simulation.runningSimulationLock.Lock()
				simulation.numAccidents += 1
				simulation.runningSimulationLock.Unlock()
				prevLocState := nextLoc.getLocationState()
				nextLoc.setLocationState(AccidentLocationState)

				currLoc.removeCar(car)
				nextLoc.addCar(car)
				go HandleAccident(&Accident{prevLocationState: prevLocState, loc: nextLoc, resolution: Unresolved, removalRate: simulation.config.carRemovalRate, probRestart: simulation.config.carRestartProb}, accidentChan)
				drawUpdateChan <- true
				break
			}

			if !(UniformRand() < simulation.config.probEnteringIntersection) { // doesn't enter intersection try again
				go MoveSmartCarInLane(car, carClock, nextLoc) // just try again with another exponential clock
				break
			}

			currLoc.removeCar(car)

			pollicePullsOver := UniformRand() < simulation.config.probPolicePullOverProb
			if simulation.config.speedBasedPullOver {
				_, prob := getNewCarSpeed(simulation.config.CarDistributionType, simulation.config.carSpeedUniformEndRange)
				if simulation.config.probPolicePullOverProb < prob {
					pollicePullsOver = true
				}
			}
			// Re sample if config is available
			if (currLoc.getLocationState() == CrossWalk || pollicePullsOver) && !car.isSlowingDown() {
				car.setSlowingDown(true)
				go HandleCrossWalkSlowCar(&SlowCar{car: car, oldSpeed: car.Speed, slowDownRate: simulation.config.crossWalkSlowDownRate}, crossWalkClock)
				car.setSpeed(simulation.config.slowDownSpeed)
			}

			if simulation.config.reSampleSpeedEveryClk && !car.slowingDown {
				speed, _ := getNewCarSpeed(simulation.config.CarDistributionType, simulation.config.carSpeedUniformEndRange)
				car.setSpeed(speed)
			}

			nextLoc.addCar(car)
			go MoveSmartCarInLane(car, carClock, nextLoc) // If next position blocked, attempt to move again on a exponential clock

			drawUpdateChan <- true
			break
		case accident := <-accidentChan:
			if !simulation.isRunningSimulation() {
				return
			}
			accident.loc.setLocationState(accident.prevLocationState)
			accident.loc.locationLock.Lock()
			cars := accident.loc.Cars
			accident.loc.locationLock.Unlock()

			if accident.resolution == Resolved {
				for _, car := range cars {
					go MoveSmartCarInLane(car, carClock, accident.loc)
				}
			} else {
				// handle removing the cars by setting them to deleted and moving them to out root
				var root *StatefulLocation
				for _, car := range cars {
					accident.loc.removeCar(car)
					if car.Direction == Horizontal {
						root = simulation.OutHorizontalRoot
					} else {
						root = simulation.OutVerticalRoot
					}

					car.carState = Deleted
					root.addCar(car)
				}
			}
			break
		case parkingCar := <-parkingChan:
			if !simulation.isRunningSimulation() {
				return
			}
			var openLanes []*StatefulLocation
			openLanes = simulation.getVerticalLanesAtIndex(parkingCar.prevLoc.X, Open)
			if len(openLanes) == 0 {
				go HandleParking(parkingCar, parkingChan) // retry bc no item in lane is free
				break
			}
			nextLoc := simulation.RandomlyPickLocation(openLanes, parkingCar.car.Direction, simulation.config.laneSwitchChoice) // TODO consider whether the car can pick its own position to switch to

			if !nextLoc.isEmpty() {
				go HandleParking(parkingCar, parkingChan)
				break
				// send the car back into parking if there is no spot to return to
			}

			parkingCar.parkingLoc.removeCar(parkingCar.car) // remove a specific car from parking
			nextLoc.addCar(parkingCar.car)
			go MoveSmartCarInLane(parkingCar.car, carClock, nextLoc) // If next position blocked, attempt to move again on a exponential clock
			simulation.RemoveCrossWalkIfNeeded(parkingCar.parkingLoc, parkingCar.car.Direction)
			break
		case slowCar := <-crossWalkClock:
			if !simulation.isRunningSimulation() {
				return
			}
			slowCar.car.setSlowingDown(false)
			slowCar.car.setSpeed(slowCar.oldSpeed)
		case <-simulation.cancelSimulation:
			simulation.setRunningSimulation(false)
			return
		}
	}
}

func (sim *GeneralLaneSimulation) AddCrossWalkIfNeeded(parkingLoc *StatefulLocation, direction Direction) {
	parkingLoc.locationLock.Lock()
	cutoff := len(parkingLoc.Cars) >= sim.config.crossWalkCutoff
	parkingLoc.locationLock.Unlock()

	if cutoff {
		var locs []*StatefulLocation
		if direction == Horizontal {
			locs = sim.getHorizontalLanesAtIndex(parkingLoc.Y, AllLocationTypes)
		} else {
			locs = sim.getVerticalLanesAtIndex(parkingLoc.X, AllLocationTypes)
		}
		for _, loc := range locs {
			if loc.getLocationState() == LaneLoc {
				loc.setLocationState(CrossWalk)
			}
		}
	}

}

func (sim *GeneralLaneSimulation) RemoveCrossWalkIfNeeded(parkingLoc *StatefulLocation, direction Direction) {
	parkingLoc.locationLock.Lock()
	cutoff := len(parkingLoc.Cars) < sim.config.crossWalkCutoff
	parkingLoc.locationLock.Unlock()

	if cutoff {
		var locs []*StatefulLocation
		if direction == Horizontal {
			locs = sim.getHorizontalLanesAtIndex(parkingLoc.Y, AllLocationTypes)
		} else {
			locs = sim.getVerticalLanesAtIndex(parkingLoc.X, AllLocationTypes)
		}
		for _, loc := range locs {
			if loc.getLocationState() == CrossWalk {
				loc.setLocationState(LaneLoc)
			}
		}
	}
}
func (sim *GeneralLaneSimulation) isCompleted() bool {
	sim.OutHorizontalRoot.locationLock.Lock()
	defer sim.OutHorizontalRoot.locationLock.Unlock()

	sim.OutVerticalRoot.locationLock.Lock()
	defer sim.OutVerticalRoot.locationLock.Unlock()

	return len(sim.OutHorizontalRoot.Cars) == sim.config.numHorizontalCars &&
		len(sim.OutVerticalRoot.Cars) == sim.config.numVerticalCars
}

func HandleCrossWalkSlowCar(slowCar *SlowCar, movementChan chan *SlowCar) {
	movementTime := rand.ExpFloat64() / slowCar.slowDownRate
	select {
	case <-time.After(time.Duration(movementTime) * time.Second):
		movementChan <- slowCar
	}
}

func HandleParking(parking *Parking, movementChan chan *Parking) {
	movementTime := rand.ExpFloat64() / parking.parkingTimeRate
	select {
	case <-time.After(time.Duration(movementTime) * time.Second):
		movementChan <- parking
	}
}

func HandleAccident(accident *Accident, movementChan chan *Accident) {
	movementTime := rand.ExpFloat64() / accident.removalRate
	select {
	case <-time.After(time.Duration(movementTime) * time.Second):
		if UniformRand() < accident.probRestart {
			accident.resolution = Resolved
			movementChan <- accident
			break
		} else {
			accident.resolution = ToBeDeleted
			movementChan <- accident
		}
	}
}

// MoveCarInLane moves the car through a lane using an exponential clock and probability of movement
func MoveSmartCarInLane(car *SmartCar, movementChan chan *SmartCar, carLoc *StatefulLocation) {
	var exponential = distuv.Exponential{Rate: car.getSpeed()}
	movementTime := exponential.Rand()
	//fmt.Println("clock started for", car.ID, car.X, car.Y)
	car.smartCarLock.Lock()
	x := car.X
	y := car.Y
	car.smartCarLock.Unlock()

	if x == -1 || y == -1 {
		return
	}
	select {
	case <-time.After(time.Duration(movementTime) * time.Second):
		car.smartCarLock.Lock()
		x := car.X
		y := car.Y
		car.smartCarLock.Unlock()

		if x == -1 || y == -1 {
			return
		}

		if carLoc.getLocationState() == AccidentLocationState {
			return // if Accident ignore the car
		}

		//fmt.Println("clock fired for", car.ID, car.X, car.Y)
		if UniformRand() < car.probMovement {
			movementChan <- car
			return
		}
		go MoveSmartCarInLane(car, movementChan, carLoc)
	}
}

func moveCarsThroughBinsDirection(
	movementChan chan Direction,
	direction Direction,
	start bool,
	location *StatefulLocation, rate float64) {
	var exponential = distuv.Exponential{Rate: rate}
	movementTime := exponential.Rand()
	timeout := time.After(time.Duration(movementTime) * time.Second)
	for {
		select {
		case <-timeout:
			movementChan <- direction
			movementTime = exponential.Rand()
			timeout = time.After(time.Duration(movementTime) * time.Second)
			break
		default:
			if start {
				location.locationLock.Lock()
				if len(location.Cars) == 0 {
					location.locationLock.Unlock()
					return
				}
				location.locationLock.Unlock()
			}
		}
	}
}
