run:
	go build  -o ./ee126_car_simulation *.go && ./ee126_car_simulation

hotreload:
	./air

build:
	go build  -o ./ee126_car_simulation *.go

.PHONY: run hotreload build