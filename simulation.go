package main

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

type Lane struct {
	Locations  []Location
	sizeOfLane int
}

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
