package main

import (
	"fmt"
	"time"
)

func RunExperiments() {
	fmt.Println("Started experiment")
	runExperimentCarDistributionTypeExperiment()
	varyingAccidentProbExperiment()
	intersectionProbabilityExperiment()
	laneAccidentProb()

	fmt.Println("Completed experiment")
}

func laneAccidentProb() {
	fmt.Println("Test varying lane Accident Prob")
	config := DefaultGeneralLaneConfig()
	config.accidentScaling = false
	config.accidentProb = 0
	config.numVerticalCars = 10
	config.numHorizontalCars = 10
	config.numVerticalLanes = 1
	config.numHorizontalLanes = 1
	config.carMovementP = 1
	config.accidentProb = 0
	config.reSampleSpeedEveryClk = false

	laneSwitchProb := []float64{0, 0.25, 0.5, 0.75, 1}
	for _, prob := range laneSwitchProb {
		config.probSwitchingLanes = prob
		fmt.Println("lane switch probability: ", prob)
		lanes := []int{1, 2, 5, 10}
		for _, lane := range lanes {
			fmt.Println("Lane size of: ", lane)
			config.numVerticalLanes = lane
			config.numHorizontalLanes = lane
			runExperiment(config)
		}

	}

}

func intersectionProbabilityExperiment() {
	fmt.Println("Test varying intersection accident probabilties")
	config := DefaultGeneralLaneConfig()
	config.accidentScaling = false
	config.accidentProb = 0
	config.numVerticalCars = 10
	config.numHorizontalCars = 10
	config.numVerticalLanes = 1
	config.numHorizontalLanes = 1
	config.carMovementP = 1
	config.accidentProb = 0
	config.reSampleSpeedEveryClk = false

	intersectionAccidentProb := []float64{0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1}
	lambas := []float64{0, 0.5, 1, 2, 5}
	for _, lambda := range lambas {
		config.carClock = lambda
		for _, prob := range intersectionAccidentProb {
			config.intersectionAccidentProb = prob
			fmt.Println("Accident probability: ", prob)
			runExperiment(config)
		}
	}
	println("Now with accident scaling")

	config.accidentScaling = true
	for _, lambda := range lambas {
		config.carClock = lambda
		for _, prob := range intersectionAccidentProb {
			config.intersectionAccidentProb = prob
			fmt.Println("Accident probability: ", prob)
			runExperiment(config)
		}
	}
}

func varyingAccidentProbExperiment() {
	fmt.Println("Test varying accident probabilities")
	config := DefaultGeneralLaneConfig()
	config.accidentScaling = false
	config.accidentProb = 0
	config.numVerticalCars = 10
	config.numHorizontalCars = 10
	config.numVerticalLanes = 1
	config.numHorizontalLanes = 1
	config.reSampleSpeedEveryClk = true
	config.carMovementP = 1
	accidentProbs := []float64{0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1}
	for _, prob := range accidentProbs {
		config.accidentProb = prob
		fmt.Println("Accident probability: ", prob)
		runExperiment(config)

	}
}
func runExperimentCarDistributionTypeExperiment() {
	fmt.Println("Test running time with different car simulations")
	config := DefaultGeneralLaneConfig()
	config.accidentScaling = false
	config.accidentProb = 0
	config.numVerticalCars = 10
	config.numHorizontalCars = 10
	config.numVerticalLanes = 1
	config.numHorizontalLanes = 1
	config.reSampleSpeedEveryClk = true
	config.carMovementP = 1
	accidentProbs := []float64{0, 0.25, 0.5, 0.75, 1}
	for _, prob := range accidentProbs {
		config.accidentProb = prob
		fmt.Println("Accident probability: ", prob)
		for i := 0; i <= 4; i++ {
			config.CarDistributionType = convertToCarDistributionType(i)
			fmt.Println("Car Distribution Type", i)
			runExperiment(config)
		}
	}
}

func runExperiment(config *GeneralLaneSimulationConfig) {
	fmt.Println("start run simulation")

	simulation, err := initMultiLaneSimulation(config)
	simulation.setRunningSimulation(true)
	if err != nil {
		// TODO handle this
	}
	start := time.Now()

	go RunGeneralSimulation(simulation)
	for {
		if !simulation.isRunningSimulation() {
			fmt.Println("General Lane Simulation completed")
			t := time.Now()

			elapsed := t.Sub(start)
			fmt.Println("Time Elapsed ", elapsed)

			simulation.runningSimulationLock.Lock()
			fmt.Println("Num Accidents ", simulation.numAccidents)
			simulation.runningSimulationLock.Unlock()
			return
		}
		select {
		case <-simulation.drawUpdateChan:
			//fmt.Println(user.simulation)
			break
		}
	}
}
