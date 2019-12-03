package main

import (
	"fmt"
	"github.com/go-siris/siris/core/errors"
	"gonum.org/v1/gonum/stat/distuv"
	"gonum.org/v1/gonum/stat/sampleuv"
	"math/rand"
	"strings"
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
}

type LocationState int

const (
	Empty LocationState = iota
	Intersection
	LaneLoc
	ParkingLoc
	AccidentLocationState
	CrossWalk
)

// Location is one spot on a lane
type StatefulLocation struct {
	Cars          map[string]*SmartCar // Allows for easy removal of the car
	LocationState LocationState
	x             int
	y             int
}

func getNewCarSpeed(speedType CarDistributionType, carSpeedEndRange float64) float64 {
	var speed float64
	if speedType == constantDistribution {
		speed = 1.0
	} else if speedType == normalDistribution {
		var UnitNormal = distuv.Normal{Mu: 0, Sigma: 1}
		speed = UnitNormal.Rand()
	} else if speedType == exponentialDistribution {
		var exponential = distuv.Exponential{Rate: 1}
		speed = exponential.Rand()
	} else if speedType == poissonDistribution {
		var poisson = distuv.Poisson{Lambda: 1}
		speed = poisson.Rand()
	} else if speedType == uniformDistribution {
		speed = UniformRandMinMax(0, carSpeedEndRange)
	}
	return speed
}

type SlowCar struct {
	car          *SmartCar
	oldSpeed     float64
	slowDownRate float64
}

func (loc *StatefulLocation) canMoveToParking() bool {
	return loc.LocationState == LaneLoc || loc.LocationState == CrossWalk || loc.LocationState == AccidentLocationState
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
		speed := getNewCarSpeed(speedType, carSpeedEndRange)
		loc.Cars[id] = &SmartCar{
			ID:        id,
			Direction: direction,
			X:         -1, Y: -1,
			Speed:        speed,
			probMovement: probMovement,
			carState:     Working}
	}
}

func (loc *StatefulLocation) isEmpty() bool {
	return len(loc.Cars) == 0 &&
		(loc.LocationState == LaneLoc ||
			loc.LocationState == Intersection ||
			loc.LocationState == CrossWalk)
}

func (loc *StatefulLocation) removeCar(car *SmartCar) {
	delete(loc.Cars, car.ID)
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
	car.X = loc.x
	car.Y = loc.y
	loc.Cars[car.ID] = car
}

type CarDistributionType int

const (
	exponentialDistribution CarDistributionType = iota
	normalDistribution      CarDistributionType = iota
	poissonDistribution     CarDistributionType = iota
	constantDistribution    CarDistributionType = iota
	uniformDistribution     CarDistributionType = iota
)

type LaneChoice int

const (
	trafficBasedChoice LaneChoice = iota
	uniformLaneChoice
)

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
	poissonProbCutoffProb float64

	carRemovalRate float64 // exponential time
	carRestartRate float64 // car restart rate

	// Handles different car rates
	carClock                float64 // for uniform, defaults to start of range
	carSpeedUniformEndRange float64
	CarDistributionType     CarDistributionType
	reSampleSpeedEveryClk   bool

	// handles cars going to fast
	carPoliceCutoff           float64
	probPolicePullOver        float64
	speedBasedAccidentScaling bool

	// parking
	parkingEnabled    bool
	distractionRate   float64 // poisson to get into parking
	parkingTimeRate   float64 // how long should car stay in parking
	parkingProbCutoff float64 // number of cars in parking
	crossWalkCutoff   int     /// numbers of cars in parking

	// crosswalk
	crossWalkEnabled      bool
	crossWalkSlowDownRate float64

	pedestrianDeathAccidentProb float64 //

	// intersection
	probEnteringIntersection float64
	intersectionAccidentRate float64 // if unspecified the same as regular accident probability
	accidentScaling          bool    // retry for accident based on the number of cars there
	slowDownSpeed            float64
	// scales poisson rate by certain amount
}

func DefaultGeneralLaneConfig() GeneralLaneSimulationConfig {
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

	config.poissonProbCutoffProb = 0

	config.carClock = 1
	config.CarDistributionType = constantDistribution
	config.reSampleSpeedEveryClk = false

	config.parkingEnabled = false
	config.probEnteringIntersection = 1
	config.parkingTimeRate = 1
	config.accidentScaling = false
	config.crossWalkCutoff = 2

	config.intersectionAccidentRate = 0
	return config
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

	carClock       chan *SmartCar
	crossWalkClock chan *SlowCar

	accidentChan chan *Accident

	parkingReturn chan *Parking
}

type JsonGeneralLocation struct {
	Cars          map[string]SmartCar `json:"cars"` // Allows for easy removal of the car
	LocationState LocationState       `json:"state"`
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
			jsonGen.Locations[i][j].LocationState = sim.Locations[i][j].LocationState
			for k, v := range sim.Locations[i][j].Cars {
				jsonGen.Locations[i][j].Cars[k] = *v
			}
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
			if len(loc.Cars) == 0 {
				fmt.Fprintf(&b, RightPad2Len("", "_", 8)+" ")
			} else {
				car := loc.getCar(false)
				fmt.Fprintf(&b, RightPad2Len(car.ID, " ", 8)+" ")
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
			locations[i][j] = &StatefulLocation{LocationState: Empty, Cars: make(map[string]*SmartCar)}
			locations[i][j].x = i
			locations[i][j].y = j
		}
	}

	// Initialize horizontal locations
	horizontalStartIndex, horizontalEndIndex := simulation.horizontalIndexRange()
	for i := horizontalStartIndex; i <= horizontalEndIndex; i++ {
		for j := 0; j < sizeOfLane; j++ {
			locations[i][j].LocationState = LaneLoc
			locations[i][j].Cars = make(map[string]*SmartCar, 0)
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
		simulation.OutHorizontalRoot = &StatefulLocation{Cars: make(map[string]*SmartCar, 0)}
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
		simulation.OutVerticalRoot = &StatefulLocation{Cars: make(map[string]*SmartCar, 0)}
	}
	// Note: Decided to only use one position for parking
	// Initialize ParkingLoc Locations
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

	simulation.moveCarsIn = make(chan Direction)
	simulation.moveCarsOut = make(chan Direction)

	simulation.carClock = make(chan *SmartCar)
	simulation.accidentChan = make(chan *Accident)
	simulation.parkingReturn = make(chan *Parking)
	simulation.crossWalkClock = make(chan *SlowCar)

	simulation.runningSimulation = false
	simulation.drawUpdateChan = make(chan bool)
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
	if location.LocationState == Intersection || location.LocationState == LaneLoc {
		return
	}
	location.LocationState = ParkingLoc
}
func isInBounds(index int, size int) bool {
	return index >= 0 && index < size
}
func (sim *GeneralLaneSimulation) countNumCarsNearby(loc *StatefulLocation) int {
	count := 1
	if isInBounds(loc.x+1, sim.config.sizeOfLane) && !sim.Locations[loc.x+1][loc.y].isEmpty() {
		count++
	}
	if isInBounds(loc.x-1, sim.config.sizeOfLane) && !sim.Locations[loc.x-1][loc.y].isEmpty() {
		count++
	}
	if isInBounds(loc.y+1, sim.config.sizeOfLane) && !sim.Locations[loc.x-1][loc.y+1].isEmpty() {
		count++
	}
	if isInBounds(loc.y-1, sim.config.sizeOfLane) && !sim.Locations[loc.x-1][loc.y-1].isEmpty() {
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
			i = item.x
		} else {
			i = item.y
		}

		count := 0.0
		for j := 0; j < sim.config.sizeOfLane; j++ {
			if direction == Horizontal && !sim.Locations[i][j].isEmpty() ||
				direction == Vertical && !sim.Locations[j][i].isEmpty() {
				count++
			}
			weights = append(weights, count)
		}

		totalCount += count
	}
	weights = normalize(weights, totalCount)

	w := sampleuv.NewWeighted(
		weights,
		nil,
	)

	i, _ = w.Take()
	return locs[i]
}

func (sim *GeneralLaneSimulation) RandomlyPickLocation(lanes []*StatefulLocation, direction Direction, choice LaneChoice) *StatefulLocation {
	if choice == uniformLaneChoice {
		return lanes[rand.Intn(len(lanes))]
	}
	return sim.selectBasedOnTraffic(lanes, direction)

}

// RunSingleLaneSimulation runs the simulation such that all the cars from bin 0 move to the last bin
func RunGeneralSimulation(simulation *GeneralLaneSimulation) {
	defer simulation.close()

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
		if len(simulation.OutHorizontalRoot.Cars) == simulation.config.numHorizontalCars &&
			len(simulation.OutVerticalRoot.Cars) == simulation.config.numVerticalCars {
			return
		}

		select {
		case carInDirection := <-moveCarsIn:
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
			go MoveSmartCarInLane(currCar, carClock, chosenLoc)
			drawUpdateChan <- true
			break
		case carOutDirection := <-moveCarsOut:
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
			drawUpdateChan <- true

			break

		case car := <-carClock:
			if car == nil {
				break
			}
			currLoc := simulation.Locations[car.X][car.Y]
			var nextLoc *StatefulLocation
			switchLanes := UniformRand() < car.probMovement

			if car.Direction == Horizontal {
				if car.Y+1 == simulation.config.sizeOfLane {
					break
				}
				if !switchLanes {
					nextLoc = simulation.Locations[car.X][car.Y+1]
				} else {
					var openLanes []*StatefulLocation
					openLanes = simulation.getHorizontalLanesAtIndex(car.Y+1, AllLocationTypes)
					nextLoc = simulation.RandomlyPickLocation(openLanes, car.Direction, simulation.config.laneSwitchChoice) // TODO consider whether the car can pick its own position to switch to
				}
			} else if car.Direction == Vertical {
				if car.X+1 == simulation.config.sizeOfLane {
					break
				}
				if !switchLanes {
					nextLoc = simulation.Locations[car.X+1][car.Y]
				} else {
					var openLanes []*StatefulLocation
					openLanes = simulation.getVerticalLanesAtIndex(car.X+1, AllLocationTypes)
					nextLoc = simulation.RandomlyPickLocation(openLanes, car.Direction, simulation.config.laneSwitchChoice)
				}
			}

			if nextLoc.LocationState == AccidentLocationState {
				go MoveSmartCarInLane(car, carClock, nextLoc) // just try again later
				break
			}

			if simulation.config.parkingEnabled && currLoc.canMoveToParking() { // parking can only happen on regular lane
				poisson := distuv.Poisson{Lambda: simulation.config.distractionRate}
				distractionOccurs := poisson.Prob(poisson.Rand()) < simulation.config.parkingProbCutoff
				if distractionOccurs {
					var parkingLoc *StatefulLocation
					if car.Direction == Horizontal {
						_, bottomEnd := simulation.horizontalIndexRange()
						if isInBounds(bottomEnd+1, simulation.config.sizeOfLane) {
							parkingLoc = simulation.Locations[car.X][bottomEnd+1]
						}
					} else {
						_, bottomEnd := simulation.verticalIndexRange()
						if isInBounds(bottomEnd+1, simulation.config.sizeOfLane) {
							parkingLoc = simulation.Locations[bottomEnd+1][car.Y]
						}
					}
					if parkingLoc != nil {
						currLoc.removeCar(car)
						parkingLoc.addCar(car)
						go HandleParking(&Parking{prevLoc: currLoc, car: car, parkingTimeRate: simulation.config.parkingTimeRate, parkingLoc: parkingLoc}, parkingChan)
						simulation.AddCrossWalkIfNeeded(parkingLoc, car.Direction)
					}
				}
			}

			var accidentOccurs = false
			if len(nextLoc.Cars) != 0 {
				poisson := distuv.Poisson{Lambda: 1}
				if nextLoc.LocationState == Intersection && simulation.config.intersectionAccidentRate != 0 {
					poisson = distuv.Poisson{Lambda: simulation.config.intersectionAccidentRate}
				}
				accidentOccurs = poisson.Prob(poisson.Rand()) < simulation.config.poissonProbCutoffProb

				if simulation.config.accidentScaling {
					numCarsNearby := simulation.countNumCarsNearby(nextLoc)
					for i := 0; i < numCarsNearby; i++ {
						if accidentOccurs {
							break
						}
						accidentOccurs = poisson.Prob(poisson.Rand()) < simulation.config.poissonProbCutoffProb
					}
				}

				// Handling of possible accident

				if !accidentOccurs {
					go MoveSmartCarInLane(car, carClock, nextLoc)
				}
				// If next position blocked, attempt to move again on a exponential clock
				break
			}

			if !accidentOccurs && nextLoc.LocationState == CrossWalk {
				poisson := distuv.Poisson{Lambda: 1}
				accidentOccurs = poisson.Prob(poisson.Rand()) < simulation.config.pedestrianDeathAccidentProb
			}

			if accidentOccurs {
				nextLoc.LocationState = AccidentLocationState
				currLoc.removeCar(car)
				nextLoc.addCar(car)
				go HandleAccident(&Accident{loc: nextLoc, resolution: Unresolved}, accidentChan)
				drawUpdateChan <- true
				break
			}

			if !(UniformRand() < simulation.config.probEnteringIntersection) { // doesn't enter intersection try again
				go MoveSmartCarInLane(car, carClock, nextLoc) // just try again with another exponential clock
				break
			}

			currLoc.removeCar(car) // TODO update this to pick only the car being referenced

			// Re sample if config is available
			if currLoc.LocationState == CrossWalk && !car.slowingDown {
				car.slowingDown = true
				go HandleCrossWalkSlowCar(&SlowCar{car: car, oldSpeed: car.Speed, slowDownRate: simulation.config.crossWalkSlowDownRate}, crossWalkClock)
				car.Speed = simulation.config.slowDownSpeed
			}
			if simulation.config.reSampleSpeedEveryClk && !car.slowingDown {
				car.Speed = getNewCarSpeed(simulation.config.CarDistributionType, simulation.config.carSpeedUniformEndRange)
			}

			nextLoc.addCar(car)
			go MoveSmartCarInLane(car, carClock, nextLoc) // If next position blocked, attempt to move again on a exponential clock

			drawUpdateChan <- true
			break
		case accident := <-accidentChan:
			accident.loc.LocationState = accident.prevLocationState
			if accident.resolution == Resolved {
				for _, car := range accident.loc.Cars {
					go MoveSmartCarInLane(car, carClock, accident.loc)
				}
			} else {
				// handle removing the cars by setting them to deleted and moving them to out root
				var root *StatefulLocation
				for _, car := range accident.loc.Cars {
					delete(accident.loc.Cars, car.ID)
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
			var openLanes []*StatefulLocation
			openLanes = simulation.getVerticalLanesAtIndex(parkingCar.prevLoc.x, Open)
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
			slowCar.car.slowingDown = false
			slowCar.car.Speed = slowCar.oldSpeed
		case <-simulation.cancelSimulation:
			return
		}
	}
}

func (sim *GeneralLaneSimulation) AddCrossWalkIfNeeded(parkingLoc *StatefulLocation, direction Direction) {
	if len(parkingLoc.Cars) >= sim.config.crossWalkCutoff {
		var locs []*StatefulLocation
		if direction == Horizontal {
			locs = sim.getHorizontalLanesAtIndex(parkingLoc.y, AllLocationTypes)
		} else {
			locs = sim.getVerticalLanesAtIndex(parkingLoc.x, AllLocationTypes)
		}
		for _, loc := range locs {
			if loc.LocationState == LaneLoc {
				loc.LocationState = CrossWalk
			}
		}
	}

}

func (sim *GeneralLaneSimulation) RemoveCrossWalkIfNeeded(parkingLoc *StatefulLocation, direction Direction) {
	if len(parkingLoc.Cars) < sim.config.crossWalkCutoff {
		var locs []*StatefulLocation
		if direction == Horizontal {
			locs = sim.getHorizontalLanesAtIndex(parkingLoc.y, AllLocationTypes)
		} else {
			locs = sim.getVerticalLanesAtIndex(parkingLoc.x, AllLocationTypes)
		}
		for _, loc := range locs {
			if loc.LocationState == CrossWalk {
				loc.LocationState = LaneLoc
			}
		}
	}
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
		accident.resolution = ToBeDeleted
		movementChan <- accident
	default:
		if UniformRand() < accident.probRestart {
			accident.resolution = Resolved
			movementChan <- accident
		}
		// handle prob of starting again
	}
}

// MoveCarInLane moves the car through a lane using an exponential clock and probability of movement
func MoveSmartCarInLane(car *SmartCar, movementChan chan *SmartCar, carLoc *StatefulLocation) {
	var exponential = distuv.Exponential{Rate: car.Speed}
	movementTime := exponential.Rand()
	select {
	case <-time.After(time.Duration(movementTime) * time.Second):
		if carLoc.LocationState == AccidentLocationState {
			return // if Accident ignore the car
		}
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
	for {
		var exponential = distuv.Exponential{Rate: rate}
		movementTime := exponential.Rand()
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
