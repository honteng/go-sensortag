package sensortag

import (
	"encoding/binary"
	"fmt"

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

func (c *CC2650) SubscribeIrTemperature(f func(float64, float64)) error {

	return c.client.Subscribe(c.chars[IR_TEMPERATURE_DATA_UUID], false, func(b []byte) {
		obj := int16(binary.LittleEndian.Uint16(b[0:]))
		amb := int16(binary.LittleEndian.Uint16(b[2:]))

		f(float64(obj)/128.0, float64(amb)/128.0)
	})
}

func (c *CC2650) EnableIrTemperature() error {
	//WriteCharacteristic(c *Characteristic, value []byte, noRsp bool) error
	return c.client.WriteCharacteristic(c.chars[IR_TEMPERATURE_CONFIG_UUID], []byte{0x01}, false)

}
