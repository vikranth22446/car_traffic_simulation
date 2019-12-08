package main

import (
	"github.com/urfave/cli"
	"io/ioutil"
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
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:     "nl",
					Usage:    "removes logging from output. Add -nl to the end of the command when running",
					Required: false,
				},
				&cli.IntFlag{
					Name:     "port",
					Usage:    "sets port",
					Value:    5000,
					Required: false,
				},
			},
			Action: func(c *cli.Context) {
				if c.Bool("nl") {
					log.SetOutput(ioutil.Discard)
				}
				runServer(c.Int("port"))
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
