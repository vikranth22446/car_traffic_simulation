
help:
	@echo "(1) make run: To build and run it";
	@echo "(2) make hotreload: To run air";
	@echo "(3) make build: To just build it";

run:
	go build  -o ./ee126_car_simulation *.go && ./ee126_car_simulation

hotreload:
	./air

build:
	go build  -o ./ee126_car_simulation *.go

.PHONY: run hotreload build