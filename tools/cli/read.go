package main

import (
	"context"
	"sync"
	"time"

	hd52_3d "github.com/ngyewch/hd52.3d"
	"github.com/simonvetter/modbus"
	"github.com/urfave/cli/v3"
	"github.com/yassinebenaid/godump"
)

var (
	dumper = godump.Dumper{
		Indentation:       "  ",
		HidePrivateFields: true,
		Theme:             godump.DefaultTheme,
	}
)

func newDev(ctx context.Context, cmd *cli.Command) (*hd52_3d.Dev, error) {
	serialPort := cmd.String(serialPortFlag.Name)
	modbusUnitId := cmd.Uint(modbusUnitIdFlag.Name)

	client, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:      "rtu://" + serialPort,
		Speed:    115200,
		DataBits: 8,
		Parity:   modbus.PARITY_EVEN,
		StopBits: 1,
		Timeout:  5 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	err = client.Open()
	if err != nil {
		return nil, err
	}

	var mutex sync.Mutex

	dev := hd52_3d.New(client, uint8(modbusUnitId), &mutex)

	return dev, nil
}

func doRead(ctx context.Context, cmd *cli.Command) error {
	e, err := newDev(ctx, cmd)
	if err != nil {
		return err
	}

	reading, err := e.Read()
	if err != nil {
		return err
	}

	return dumper.Println(reading)
}
