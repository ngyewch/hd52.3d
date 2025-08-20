package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime/debug"

	"github.com/urfave/cli/v3"
)

var (
	serialPortFlag = &cli.StringFlag{
		Name:     "serial-port",
		Usage:    "serial port",
		Required: true,
		Sources:  cli.EnvVars("SERIAL_PORT"),
	}
	modbusUnitIdFlag = &cli.UintFlag{
		Name:    "modbus-unit-id",
		Usage:   "ModBus unit ID",
		Value:   1,
		Sources: cli.EnvVars("MODBUS_UNIT_ID"),
		Action: func(ctx context.Context, cmd *cli.Command, v uint) error {
			if (v < 1) || (v > 247) {
				return fmt.Errorf("invalid modbus-unit-id: %d", v)
			}
			return nil
		},
	}

	app = &cli.Command{
		Name:  "hd52.3d",
		Usage: "DeltaOHM 52.3d CLI",
		Flags: []cli.Flag{
			serialPortFlag,
			modbusUnitIdFlag,
		},
		Commands: []*cli.Command{
			{
				Name:   "read",
				Usage:  "doRead",
				Action: doRead,
			},
		},
	}
)

func main() {
	buildInfo, _ := debug.ReadBuildInfo()
	if buildInfo != nil {
		app.Version = buildInfo.Main.Version
	}

	err := app.Run(context.Background(), os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
