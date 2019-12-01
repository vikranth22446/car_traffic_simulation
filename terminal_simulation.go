package main

import (
	"fmt"
	"time"
)

// RunTerminalSingleLaneSimulation simulates and prints the movement through a single lane
func RunTerminalSingleLaneSimulation(action bool) {
	simulation := initSingleLaneSimulation(10)
	simulation.runningSimulation = true

	go RunSingleLaneSimulation(simulation)
	if action {
		for {
			if !simulation.runningSimulation {
				fmt.Println("SingleLaneSimulation completed")
				return
			}
			select {
			case <-simulation.drawUpdateChan:
				fmt.Println(simulation)
				break
			}
		}
		return
	}

	RenderTerminalFPS(simulation)
}

// RenderTerminalFPS renders date to the screen after every fps. It may clutter up the screen and lag
func RenderTerminalFPS(simulation *SingleLaneSimulation) {
	var _, renderStartTime, diff, sleep int64
	_ = time.Now().UnixNano()
	for {

		renderStartTime = time.Now().UnixNano()

		fmt.Println(simulation.String())

		diff = time.Now().UnixNano() - renderStartTime
		sleep = fpsn - diff

		if sleep < 0 {
			continue
		}
		time.Sleep(time.Duration(sleep) * time.Nanosecond)

	}
}
