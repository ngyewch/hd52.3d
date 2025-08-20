package hd52_3d

import (
	"encoding/binary"
	"fmt"
	"sync"

	"github.com/simonvetter/modbus"
)

type Dev struct {
	mc     *modbus.ModbusClient
	unitId uint8
	mutex  *sync.Mutex
}

type WindSpeedUnit uint16

const (
	WindSpeedUnitMetersPerSecond      WindSpeedUnit = 0
	WindSpeedUnitCentimetersPerSecond WindSpeedUnit = 1
	WindSpeedUnitKilometersPerHour    WindSpeedUnit = 2
	WindSpeedUnitKnot                 WindSpeedUnit = 3
	WindSpeedUnitMilesPerHour         WindSpeedUnit = 4
)

type TemperatureUnit uint16

const (
	TemperatureUnitCelsius    TemperatureUnit = 0
	TemperatureUnitFahrenheit TemperatureUnit = 1
)

type AtmosphericPressureUnit uint16

const (
	AtmosphericPressureUnitMbar    AtmosphericPressureUnit = 0
	AtmosphericPressureUnitMmHg    AtmosphericPressureUnit = 1
	AtmosphericPressureUnitInchHg  AtmosphericPressureUnit = 2
	AtmosphericPressureUnitMmH2O   AtmosphericPressureUnit = 3
	AtmosphericPressureUnitInchH2O AtmosphericPressureUnit = 4
	AtmosphericPressureUnitAtm     AtmosphericPressureUnit = 5
)

type RainfallUnit uint16

const (
	RainfallUnitMm   RainfallUnit = 0
	RainfallUnitInch RainfallUnit = 1
)

type Reading struct {
	WindSpeed                      float64 // units: (configured)
	WindDirection                  float64 // units: degrees
	SonicTemperature1              float64 // units: (configured)
	SonicTemperature2              float64 // units: (configured)
	SonicTemperatureAveraged       float64 // units: (configured)
	Pt100Temperature               float64 // units: (configured)
	RelativeHumidity               float64 // units: %RH
	BarometricPressure             float64 // units: (configured)
	CompassAngle                   float64 // units: degrees
	SolarRadiation                 float64 // units: W/m2
	MeanWindSpeed                  float64 // units: (configured)
	MeanWindDirection              float64 // units: degrees
	AbsoluteHumidity               float64 // units: g/m3
	DewPointTemperature            float64 // units: (configured)
	WindDirectionExtended          float64 // units: degrees
	WindSpeedVAxis                 float64 // units: ?
	WindSpeedUAxis                 float64 // units: ?
	SpeedMeasurementError          bool
	CompassMeasurementError        bool
	TemperatureMeasurementError    bool
	HumidityMeasurementError       bool
	PressureMeasurementError       bool
	SolarRadiationMeasurementError bool
	WindSpeedUnit                  WindSpeedUnit
	TemperatureUnit                TemperatureUnit
	AtmosphericPressureUnit        AtmosphericPressureUnit
	WindGustIntensity              float64 // units: (configured)
	WindGustDirection              float64 // units: degrees
	TotalAmountOfRainfall          float64 // units: (configured)
	PartialAmountOfRainfall        float64 // units: (configured)
	RainfallRate                   float64 // units: (configured) / hour
	RainfallUnit                   RainfallUnit
}

func New(mc *modbus.ModbusClient, unitId uint8, mutex *sync.Mutex) *Dev {
	return &Dev{
		mc:     mc,
		unitId: unitId,
		mutex:  mutex,
	}
}

func (dev *Dev) requestSetup() error {
	err := dev.mc.SetUnitId(dev.unitId)
	if err != nil {
		return err
	}
	err = dev.mc.SetEncoding(modbus.BIG_ENDIAN, modbus.HIGH_WORD_FIRST)
	if err != nil {
		return err
	}
	return nil
}

func (dev *Dev) Read() (*Reading, error) {
	dev.mutex.Lock()
	defer dev.mutex.Unlock()

	err := dev.requestSetup()
	if err != nil {
		return nil, err
	}

	regs, err := dev.mc.ReadRegisters(0, 29, modbus.INPUT_REGISTER)
	if err != nil {
		return nil, err
	}

	status := regs[17]
	reading := Reading{
		WindSpeed:                      uint16ToFloat64(regs[0], 100),
		WindDirection:                  uint16ToFloat64(regs[1], 10),
		SonicTemperature1:              int16ToFloat64(regs[2], 10),
		SonicTemperature2:              int16ToFloat64(regs[3], 10),
		SonicTemperatureAveraged:       int16ToFloat64(regs[4], 10),
		Pt100Temperature:               int16ToFloat64(regs[5], 10),
		RelativeHumidity:               uint16ToFloat64(regs[6], 10),
		CompassAngle:                   uint16ToFloat64(regs[8], 10),
		SolarRadiation:                 uint16ToFloat64(regs[9], 1),
		MeanWindSpeed:                  uint16ToFloat64(regs[10], 100),
		MeanWindDirection:              uint16ToFloat64(regs[11], 10),
		AbsoluteHumidity:               uint16ToFloat64(regs[12], 100),
		DewPointTemperature:            int16ToFloat64(regs[13], 10),
		WindDirectionExtended:          uint16ToFloat64(regs[14], 10),
		WindSpeedVAxis:                 uint16ToFloat64(regs[15], 100),
		WindSpeedUAxis:                 uint16ToFloat64(regs[16], 100),
		SpeedMeasurementError:          bitsToBool(status, 0),
		CompassMeasurementError:        bitsToBool(status, 1),
		TemperatureMeasurementError:    bitsToBool(status, 2),
		HumidityMeasurementError:       bitsToBool(status, 3),
		PressureMeasurementError:       bitsToBool(status, 4),
		SolarRadiationMeasurementError: bitsToBool(status, 5),
		WindSpeedUnit:                  WindSpeedUnit(regs[18]),
		TemperatureUnit:                TemperatureUnit(regs[19]),
		AtmosphericPressureUnit:        AtmosphericPressureUnit(regs[20]),
		WindGustIntensity:              uint16ToFloat64(regs[21], 100),
		WindGustDirection:              uint16ToFloat64(regs[22], 10),
		RainfallUnit:                   RainfallUnit(regs[28]),
	}

	if reading.AtmosphericPressureUnit == AtmosphericPressureUnitAtm {
		reading.BarometricPressure = uint16ToFloat64(regs[7], 1000)
	} else {
		reading.BarometricPressure = uint16ToFloat64(regs[7], 10)
	}

	switch reading.RainfallUnit {
	case RainfallUnitMm:
		reading.TotalAmountOfRainfall = uint32ToFloat64(regs[23], regs[24], 1000)
		reading.PartialAmountOfRainfall = uint32ToFloat64(regs[25], regs[26], 1000)
		reading.RainfallRate = uint16ToFloat64(regs[27], 10)
	case RainfallUnitInch:
		reading.TotalAmountOfRainfall = uint32ToFloat64(regs[23], regs[24], 10000)
		reading.PartialAmountOfRainfall = uint32ToFloat64(regs[25], regs[26], 10000)
		reading.RainfallRate = uint16ToFloat64(regs[27], 100)
	default:
		return nil, fmt.Errorf("unknown rainfall unit")
	}

	return &reading, nil
}

func uint16ToFloat64(v uint16, scale float64) float64 {
	return float64(v) / scale
}

func uint32ToFloat64(msb uint16, lsb uint16, scale float64) float64 {
	var b []byte
	b = binary.BigEndian.AppendUint16(b, msb)
	b = binary.BigEndian.AppendUint16(b, lsb)
	v := binary.BigEndian.Uint32(b)
	return float64(v) / scale
}

func int16ToFloat64(v uint16, scale float64) float64 {
	return float64(int16(v)) / scale
}

func bitsToBool(v uint16, bit uint8) bool {
	return ((v >> bit) & 1) == 1
}
