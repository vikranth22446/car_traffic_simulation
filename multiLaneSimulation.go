package main
// GeneralLaneSimulation handles a general simulation with n horizontal lanes and n vertical lanes and intersections
type GeneralLaneSimulation struct {
	Simulation
	HorizontalLanes *[]Lane
	VerticalLanes   *[]Lane

	moveCarsInLane  chan *Lane
	carClock        chan *Car
	moveCarsEndLane chan *Lane
}
/**
TODO: Decide whether to use 2d array to two lists to represent the board
- 2 lanes would involve making the locations that intersect point to the same object
	- difficulty is initializing and coordinating between the roads
- 2d array would be easy to initialize, but we have to place the lanes in the right spot and keep track of information about the lanes
	- The size of the lanes also cannot be variable but constant
 */