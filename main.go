package main

import (
	"github.com/urfave/cli"
	"log"
	"os"
)

func info(app *cli.App) {
	app.Name = "Simulation CLI"
	//app.Usage = "An example CLI for ordering pizza's"
	//app.Author = "Jeroenouw"
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
