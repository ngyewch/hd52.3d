package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"runtime/debug"
)

var (
	serialPortFlag = &cli.StringFlag{
		Name:     "serial-port",
		Usage:    "serial port",
		Required: true,
		EnvVars:  []string{"SERIAL_PORT"},
	}
	modbusUnitIdFlag = &cli.UintFlag{
		Name:    "modbus-unit-id",
		Usage:   "ModBus unit ID",
		Value:   1,
		EnvVars: []string{"MODBUS_UNIT_ID"},
		Action: func(cCtx *cli.Context, v uint) error {
			if (v < 1) || (v > 247) {
				return fmt.Errorf("invalid modbus-unit-id: %d", v)
			}
			return nil
		},
	}

	app = &cli.App{
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

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
