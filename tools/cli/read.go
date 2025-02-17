package main

import (
	hd52_3d "github.com/ngyewch/hd52.3d"
	"github.com/simonvetter/modbus"
	"github.com/urfave/cli/v2"
	"github.com/yassinebenaid/godump"
	"sync"
	"time"
)

var (
	dumper = godump.Dumper{
		Indentation:       "  ",
		HidePrivateFields: true,
		Theme:             godump.DefaultTheme,
	}
)

func newDev(cCtx *cli.Context) (*hd52_3d.Dev, error) {
	serialPort := serialPortFlag.Get(cCtx)
	modbusUnitId := modbusUnitIdFlag.Get(cCtx)

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

func doRead(cCtx *cli.Context) error {
	e, err := newDev(cCtx)
	if err != nil {
		return err
	}

	reading, err := e.Read()
	if err != nil {
		return err
	}

	return dumper.Println(reading)
}
