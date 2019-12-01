package main

import (
	"fmt"
	"time"
)

func RunTerminalSimulation(action bool) {
	simulation := initSimulation(10)
	simulation.runningSimulation = true

	go RunSimulation(simulation)
	if action {
		for {
			if !simulation.runningSimulation {
				fmt.Println("Simulation completed")
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

func RenderTerminalFPS(simulation *Simulation) {
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