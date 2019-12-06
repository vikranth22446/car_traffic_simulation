package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"math/rand"
	"os"
	"strconv"
	"time"
)

// UserGroup keeps track of the current weboscket connections
type UserGroup struct {
	connUserMap map[*websocket.Conn]*User
	users       map[*User]bool
	ids         map[uuid.UUID]*User
	rules       map[string]Handler
}

func newUserGroup() *UserGroup {
	self := &UserGroup{}
	self.users = map[*User]bool{}
	self.rules = map[string]Handler{}
	self.connUserMap = map[*websocket.Conn]*User{}
	return self
}

// FindHandler implements a handler finding function for router.
func (userGroup *UserGroup) FindHandler(event string) (Handler, bool) {
	handler, found := userGroup.rules[event]
	return handler, found
}

// AddEventHandler is a function to add handlers to the router.
func (userGroup *UserGroup) AddEventHandler(event string, handler Handler) {
	userGroup.rules[event] = handler
}

func (userGroup *UserGroup) removePlayer(p *User) {
	delete(userGroup.users, p)
	delete(userGroup.ids, p.ID)
	delete(userGroup.connUserMap, p.ws)
	p.group = nil
}

var (
	userGroup *UserGroup
)

func init() {
	userGroup = newUserGroup()
	if os.Getenv("DEBUG") != "" {
		playerShowLog = true
	}
	userGroup.AddEventHandler("startSimulation", startSimulationEvent)
	userGroup.AddEventHandler("cancelSimulation", cancelSimulation)
	rand.Seed(time.Now().Unix())
}

func (userGroup *UserGroup) addUser(p *User) {
	p.group = userGroup
	newUUID, err := uuid.NewUUID()
	if err != nil {
		return
	}
	p.ID = newUUID
	userGroup.users[p] = true
	userGroup.connUserMap[p.ws] = p
}

func cancelSimulation(conn *websocket.Conn, data interface{}) {
	fmt.Println("stopping current simulation")

	user, exists := userGroup.connUserMap[conn]
	if !exists {
		return
	}
	if !user.runningSimulation {
		return
	}
	user.simulation.cancelSimulation <- true
	user.runningSimulation = false
	user.simulation.setRunningSimulation(false)
}

func startSimulationEvent(conn *websocket.Conn, data interface{}) {
	fmt.Println("start Simulation event handler")

	user, exists := userGroup.connUserMap[conn]
	if !exists {
		return
	}
	config := DefaultGeneralLaneConfig()

	m := data.(map[string]interface{})["data"].(map[string]interface{})
	//fmt.Println(m)
	if sizeOfLane, ok := m["sizeOfLane"].(string); ok {
		sizeOfLaneInt, err := strconv.Atoi(sizeOfLane)
		if err != nil {
			return
		}
		config.sizeOfLane = sizeOfLaneInt
	}

	if numHorizontalLanes, ok := m["numHorizontalLanes"].(string); ok {
		numHorizontalLanesInt, err := strconv.Atoi(numHorizontalLanes)
		if err != nil {
			return
		}
		config.numHorizontalLanes = numHorizontalLanesInt
	}

	if numVerticalLanes, ok := m["numVerticalLanes"].(string); ok {
		numVerticalLanesInt, err := strconv.Atoi(numVerticalLanes)
		if err != nil {
			return
		}
		config.numVerticalLanes = numVerticalLanesInt
	}
	if inAlpha, ok := m["inAlpha"].(string); ok {
		inAlpha, err := strconv.ParseFloat(inAlpha, 64)
		if err != nil {
			return
		}
		config.inAlpha = inAlpha
	}

	if outBeta, ok := m["outBeta"].(string); ok {
		outBeta, err := strconv.ParseFloat(outBeta, 64)
		if err != nil {
			return
		}
		config.outBeta = outBeta
	}

	if carMovementP, ok := m["carMovementP"].(string); ok {
		carMovementP, err := strconv.ParseFloat(carMovementP, 64)
		if err != nil {
			return
		}
		config.carMovementP = carMovementP
	}

	if probSwitchingLanes, ok := m["probSwitchingLanes"].(string); ok {
		probSwitchingLanes, err := strconv.ParseFloat(probSwitchingLanes, 64)
		if err != nil {
			return
		}
		config.probSwitchingLanes = probSwitchingLanes
	}

	if accidentProb, ok := m["accidentProb"].(string); ok {
		accidentProb, err := strconv.ParseFloat(accidentProb, 64)
		if err != nil {
			return
		}
		config.accidentProb = accidentProb
	}

	if numHorizontalCars, ok := m["numHorizontalCars"].(string); ok {
		numHorizontalCars, err := strconv.Atoi(numHorizontalCars)
		if err != nil {
			return
		}
		config.numHorizontalCars = numHorizontalCars
	}

	if numVerticalCars, ok := m["numVerticalCars"].(string); ok {
		numVerticalCars, err := strconv.Atoi(numVerticalCars)
		if err != nil {
			return
		}
		config.numVerticalCars = numVerticalCars
	}

	if inLaneChoice, ok := m["inLaneChoice"].(float64); ok {
		//inLaneChoice, err := strconv.Atoi(inLaneChoice)
		//if err != nil {
		//	return
		//}
		config.inLaneChoice = convertIntLaneChoice(int(inLaneChoice))
	}

	if outLaneChoice, ok := m["outLaneChoice"].(float64); ok {
		//outLaneChoice, err := strconv.Atoi(outLaneChoice)
		//if err != nil {
		//	return
		//}
		config.outLaneChoice = convertIntLaneChoice(int(outLaneChoice))
	}

	if laneSwitchChoice, ok := m["laneSwitchChoice"].(float64); ok {
		//laneSwitchChoice, err := strconv.Atoi(laneSwitchChoice)
		//if err != nil {
		//	return
		//}
		config.laneSwitchChoice = convertIntLaneChoice(int(laneSwitchChoice))
	}

	if carRemovalRate, ok := m["carRemovalRate"].(string); ok {
		carRemovalRate, err := strconv.ParseFloat(carRemovalRate, 64)
		if err != nil {
			return
		}
		config.carRemovalRate = carRemovalRate
	}

	if carRestartRate, ok := m["carRestartProb"].(string); ok {
		carRestartRate, err := strconv.ParseFloat(carRestartRate, 64)
		if err != nil {
			return
		}
		config.carRestartProb = carRestartRate
	}

	if carClock, ok := m["carClock"].(string); ok {
		carClock, err := strconv.ParseFloat(carClock, 64)
		if err != nil {
			return
		}
		config.carClock = carClock
	}

	if carSpeedUniformEndRange, ok := m["carSpeedUniformEndRange"].(string); ok {
		carSpeedUniformEndRange, err := strconv.ParseFloat(carSpeedUniformEndRange, 64)
		if err != nil {
			return
		}
		config.carSpeedUniformEndRange = carSpeedUniformEndRange
	}

	if CarDistributionType, ok := m["CarDistributionType"].(float64); ok {
		//CarDistributionType, err := strconv.Atoi(CarDistributionType)
		//if err != nil {
		//	return
		//}
		config.CarDistributionType = convertToCarDistributionType(int(CarDistributionType))
	}

	if reSampleSpeedEveryClk, ok := m["reSampleSpeedEveryClk"].(bool); ok {
		config.reSampleSpeedEveryClk = reSampleSpeedEveryClk

	}

	if probPolicePullOverProb, ok := m["probPolicePullOverProb"].(string); ok {
		probPolicePullOverProb, err := strconv.ParseFloat(probPolicePullOverProb, 64)
		if err != nil {
			return
		}
		config.probPolicePullOverProb = probPolicePullOverProb
	}

	if speedBasedPullOver, ok := m["speedBasedPullOver"].(bool); ok {
		config.speedBasedPullOver = speedBasedPullOver
	}

	if parkingEnabled, ok := m["parkingEnabled"].(bool); ok {
		config.parkingEnabled = parkingEnabled
	}

	if distractionRate, ok := m["distractionRate"].(string); ok {
		distractionRate, err := strconv.ParseFloat(distractionRate, 64)
		if err != nil {
			return
		}
		config.distractionRate = distractionRate
	}

	if parkingTimeRate, ok := m["parkingTimeRate"].(string); ok {
		parkingTimeRate, err := strconv.ParseFloat(parkingTimeRate, 64)
		if err != nil {
			return
		}
		config.parkingTimeRate = parkingTimeRate
	}

	if crossWalkCutoff, ok := m["crossWalkCutoff"].(string); ok {
		crossWalkCutoff, err := strconv.Atoi(crossWalkCutoff)
		if err != nil {
			return
		}
		config.crossWalkCutoff = crossWalkCutoff
	}

	if crossWalkEnabled, ok := m["crossWalkEnabled"].(bool); ok {
		config.crossWalkEnabled = crossWalkEnabled
	}

	if pedestrianDeathAccidentProb, ok := m["pedestrianDeathAccidentProb"].(string); ok {
		pedestrianDeathAccidentProb, err := strconv.ParseFloat(pedestrianDeathAccidentProb, 64)
		if err != nil {
			return
		}
		config.pedestrianDeathAccidentProb = pedestrianDeathAccidentProb
	}

	if probEnteringIntersection, ok := m["probEnteringIntersection"].(string); ok {
		probEnteringIntersection, err := strconv.ParseFloat(probEnteringIntersection, 64)
		if err != nil {
			return
		}
		config.probEnteringIntersection = probEnteringIntersection
	}

	if intersectionAccidentProb, ok := m["intersectionAccidentProb"].(string); ok {
		intersectionAccidentProb, err := strconv.ParseFloat(intersectionAccidentProb, 64)
		if err != nil {
			return
		}
		config.intersectionAccidentProb = intersectionAccidentProb
	}

	if accidentScaling, ok := m["accidentScaling"].(bool); ok {
		config.accidentScaling = accidentScaling
	}

	if slowDownSpeed, ok := m["slowDownSpeed"].(string); ok {
		slowDownSpeed, err := strconv.ParseFloat(slowDownSpeed, 64)
		if err != nil {
			return
		}
		config.slowDownSpeed = slowDownSpeed
	}

	user.runSimulation(config)

}
