package sensortag

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/go-ble/ble"
)

const (
	IR_TEMPERATURE_UUID      = "f000aa0004514000b000000000000000"
	HUMIDITY_UUID            = "f000aa2004514000b000000000000000"
	BAROMETRIC_PRESSURE_UUID = "f000aa4004514000b000000000000000"
	SIMPLE_KEY_UUID          = "ffe0"

	IR_TEMPERATURE_CONFIG_UUID = "f000aa0204514000b000000000000000"
	IR_TEMPERATURE_DATA_UUID   = "f000aa0104514000b000000000000000"
	IR_TEMPERATURE_PERIOD_UUID = "f000aa0304514000b000000000000000"

	HUMIDITY_CONFIG_UUID = "f000aa2204514000b000000000000000"
	HUMIDITY_DATA_UUID   = "f000aa2104514000b000000000000000"
	HUMIDITY_PERIOD_UUID = "f000aa2304514000b000000000000000"

	BAROMETRIC_PRESSURE_CONFIG_UUID = "f000aa4204514000b000000000000000"
	BAROMETRIC_PRESSURE_DATA_UUID   = "f000aa4104514000b000000000000000"
	BAROMETRIC_PRESSURE_PERIOD_UUID = "f000aa4404514000b000000000000000"

	SIMPLE_KEY_DATA_UUID = "ffe1"

	MPU9250_UUID   = "f000aa8004514000b000000000000000"
	IO_UUID        = "f000aa6404514000b000000000000000"
	LUXOMETER_UUID = "f000aa7004514000b000000000000000"

	MPU9250_CONFIG_UUID = "f000aa8204514000b000000000000000"
	MPU9250_DATA_UUID   = "f000aa8104514000b000000000000000"
	MPU9250_PERIOD_UUID = "f000aa8304514000b000000000000000"

	MPU9250_GYROSCOPE_MASK     = 0x0007
	MPU9250_ACCELEROMETER_MASK = 0x0238
	MPU9250_MAGNETOMETER_MASK  = 0x0040

	IO_DATA_UUID   = "f000aa6504514000b000000000000000"
	IO_CONFIG_UUID = "f000aa6604514000b000000000000000"

	LUXOMETER_CONFIG_UUID = "f000aa7204514000b000000000000000"
	LUXOMETER_DATA_UUID   = "f000aa7104514000b000000000000000"
	LUXOMETER_PERIOD_UUID = "f000aa7304514000b000000000000000"
)

type (
	Callbacks interface {
		OnIrTemperatureChange()
		OnHumidityChange()
		OnBarometricPressureChange()
		OnSimpleKeyChange()
	}

	CC2650 struct {
		profile *ble.Profile
		client  ble.Client
		chars   map[string]*ble.Characteristic
	}
)

func Filter(a ble.Advertisement) bool {
	localName := a.LocalName()
	fmt.Println(localName)
	return localName == "CC2650 SensorTag" || localName == "SensorTag 2.0"
}

func NewCC2650(c ble.Client, p *ble.Profile) (*CC2650, error) {
	uuids := []string{
		// common
		IR_TEMPERATURE_UUID,
		HUMIDITY_UUID,
		BAROMETRIC_PRESSURE_UUID,
		SIMPLE_KEY_UUID,
		IR_TEMPERATURE_CONFIG_UUID,
		IR_TEMPERATURE_DATA_UUID,
		IR_TEMPERATURE_PERIOD_UUID,
		HUMIDITY_CONFIG_UUID,
		HUMIDITY_DATA_UUID,
		HUMIDITY_PERIOD_UUID,
		BAROMETRIC_PRESSURE_CONFIG_UUID,
		BAROMETRIC_PRESSURE_DATA_UUID,
		BAROMETRIC_PRESSURE_PERIOD_UUID,
		SIMPLE_KEY_DATA_UUID,

		// cc2650
		MPU9250_UUID,
		IO_UUID,
		LUXOMETER_UUID,

		MPU9250_CONFIG_UUID,
		MPU9250_DATA_UUID,
		MPU9250_PERIOD_UUID,

		IO_DATA_UUID,
		IO_CONFIG_UUID,

		LUXOMETER_CONFIG_UUID,
		LUXOMETER_DATA_UUID,
		LUXOMETER_PERIOD_UUID,

		// battery

	}

	chars := make(map[string]*ble.Characteristic)

	for _, uuid := range uuids {
		chars[uuid] = nil
	}

	cnt := 0
	for _, s := range p.Services {
		fmt.Println(s.UUID.String())

		for _, c := range s.Characteristics {
			if old, ok := chars[c.UUID.String()]; ok {
				if old == nil {
					chars[c.UUID.String()] = c
					fmt.Println(c.UUID.String())
					cnt++
				} else {
					return nil, fmt.Errorf("Repetitive characteristic: %v", c.UUID)
				}
			} else {
				fmt.Printf("unknown %s\n", c.UUID)
			}
		}
	}

	// TODO: need to check the valid UUIDs?
	// 14 vs 10, diff 4 is for service?
	//if cnt != len(uuids) {
	//	return nil, fmt.Errorf("Wrong number of UUIDs(%d, %d)", cnt, len(uuids))
	//}

	fmt.Println("Succeeded to discover cc2650")
	return &CC2650{
		profile: p,
		client:  c,
		chars:   chars,
	}, nil

}

func (c *CC2650) EnableIrTemperature() error {
	return c.client.WriteCharacteristic(c.chars[IR_TEMPERATURE_CONFIG_UUID], []byte{0x01}, false)
}

func (c *CC2650) SubscribeIrTemperature(f func(float64, float64)) error {
	return c.client.Subscribe(c.chars[IR_TEMPERATURE_DATA_UUID], false, func(b []byte) {
		obj := int16(binary.LittleEndian.Uint16(b[0:]))
		amb := int16(binary.LittleEndian.Uint16(b[2:]))

		f(float64(obj)/128.0, float64(amb)/128.0)
	})
}

func (c *CC2650) UnsubscribeIrTemperature() error {
	return c.client.Unsubscribe(c.chars[IR_TEMPERATURE_DATA_UUID], false)
}

func (c *CC2650) EnableHumidity() error {
	return c.client.WriteCharacteristic(c.chars[HUMIDITY_CONFIG_UUID], []byte{0x01}, false)
}

func (c *CC2650) SubscribeHumidity(f func(float64, float64)) error {
	return c.client.Subscribe(c.chars[HUMIDITY_DATA_UUID], false, func(b []byte) {
		temp := int16(binary.LittleEndian.Uint16(b[0:]))
		hmd := int16(binary.LittleEndian.Uint16(b[2:]))

		temperature := -40 + ((165 * float64(temp)) / 65536.0)
		humidity := float64(hmd) * 100 / 65536.0

		f(humidity, temperature)
	})
}

func (c *CC2650) UnsubscribeHumidity() error {
	return c.client.Unsubscribe(c.chars[HUMIDITY_DATA_UUID], false)
}

func (c *CC2650) EnablePressure() error {
	return c.client.WriteCharacteristic(c.chars[BAROMETRIC_PRESSURE_CONFIG_UUID], []byte{0x01}, false)
}

func (c *CC2650) SubscribePressure(f func(float64, float64)) error {
	return c.client.Subscribe(c.chars[BAROMETRIC_PRESSURE_DATA_UUID], false, func(b []byte) {

		// data is returned as
		// Firmare 0.89 16 bit single precision float
		// Firmare 1.01 24 bit single precision float

		var tempBMP, pressure float64

		if len(b) > 4 {
			// Firmware 1.01
			temp := binary.LittleEndian.Uint32(b[0:])
			press := binary.LittleEndian.Uint32(b[2:])

			tempBMP = float64(temp&0x00ffffff) / 100.0
			pressure = float64((press>>8)&0x00ffffff) / 100.0
		} else {
			// Firmware 0.89
			temp := binary.LittleEndian.Uint16(b[0:])
			press := binary.LittleEndian.Uint16(b[2:])

			tempExponent := float64((temp & 0xF000) >> 12)
			tempMantissa := float64((temp & 0x0FFF))
			tempBMP = tempMantissa * math.Pow(2, tempExponent) / 100.0

			pressureExponent := float64((press & 0xF000) >> 12)
			pressureMantissa := float64(press & 0x0FFF)
			pressure = pressureMantissa * math.Pow(2, pressureExponent) / 100.0
		}

		f(tempBMP, pressure)
	})
}

func (c *CC2650) UnsubscribePresure() error {
	return c.client.Unsubscribe(c.chars[BAROMETRIC_PRESSURE_DATA_UUID], false)
}

func (c *CC2650) SubscribeSimpleKey(f func(bool)) error {
	return c.client.Subscribe(c.chars[SIMPLE_KEY_DATA_UUID], false, func(b []byte) {
		if b[0] == 1 {
			f(true)
		} else {
			f(false)
		}
	})
}

func (c *CC2650) UnsubscribeSimpleKey() error {
	return c.client.Unsubscribe(c.chars[SIMPLE_KEY_DATA_UUID], false)
}
