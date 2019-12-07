package main

import (
	"github.com/urfave/cli"
	"log"
	"os"
)

func info(app *cli.App) {
	app.Name = "SingleLaneSimulation CLI"
	app.Usage = "Simulating car traffic over time"
	app.Version = "0.0.1"
}

func commands(app *cli.App) {
	app.Commands = []cli.Command{
		{
			Name:    "start-server",
			Aliases: []string{"start"},
			Usage:   "Starts the golang server",
			Action: func(c *cli.Context) {
				runServer()
			},
		},
		{
			Name:    "run-terminal-simulation",
			Aliases: []string{"single-t"},
			Usage:   "Can run the simulation as terminal printouts",
			Action: func(c *cli.Context) {
				RunTerminalSingleLaneSimulation(true)
			},
		},
		{
			Name:    "General multi lane",
			Aliases: []string{"t"},
			Usage:   "Can run the simulation as terminal printouts",
			Action: func(c *cli.Context) {
				RunTerminalMultiLaneSimulation()
			},
		},
		{
			Name:    "Run Experiments",
			Aliases: []string{"e", "experiment"},
			Usage:   "Run the four experiments we defined in the report",
			Action: func(c *cli.Context) {
				RunExperiments()
			},
		},
	}
}

func main() {
	var app = cli.NewApp()
	info(app)
	commands(app)

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
