# Installation
Install golang using https://golang.org/doc/install

## Installing Go Dependencies
```
go get
```
## Running Code via Make
the commands make supports. Type `make` to view this description
```sh
(1) make run: To build and run it
(2) make hotreload: To run air
(3) make build: To just build it
```
## Running the CLI
After building the code, you can run a cli to start the server or run commands.
Running ee126_car_simulation executable gets the following description.
```sh
NAME:
   Simulation CLI - Simulating car traffic over time

USAGE:
   ee126_car_simulation [global options] command [command options] [arguments...]

VERSION:
   0.0.1

COMMANDS:
   start-server, start         Starts the golang server
   run-terminal-simulation, t  Can run the simulation as terminal printouts
   help, h                     Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```

## Installing Air for Hot Reload
for development on a hot reloading server,  we use the air library.
### macOS
```sh
curl -fLo ~/.air \
    https://raw.githubusercontent.com/cosmtrek/air/master/bin/darwin/air
chmod +x ~/.air
```
### Linux
```sh
curl -fLo ~/.air \
    https://raw.githubusercontent.com/cosmtrek/air/master/bin/linux/air
chmod +x ~/.air
```

### Windows
```sh
curl -fLo ~/.air.exe \
    https://raw.githubusercontent.com/cosmtrek/air/master/bin/windows/air.exe
```

# Running the Code
The command air will build the code and run it.
```
air
```

# Running the Frontend

## Installing the deps
Run `yarn install` inside the frontend directory

## Running the code with live reload
Open a tab and run `yarn start` to have it live reload

## Production deploy
Deploy for production via `yarn build`
