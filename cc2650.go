package sensortag

import "encoding/binary"

const (
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

func (c *CC2650) EnableLuxometer() error {
	return c.client.WriteCharacteristic(c.chars[LUXOMETER_CONFIG_UUID], []byte{0x01}, false)
}

func (c *CC2650) SubscribeLuxometer(f func(float64, float64)) error {
	return c.client.Subscribe(c.chars[LUXOMETER_DATA_UUID], false, func(b []byte) {
		obj := int16(binary.LittleEndian.Uint16(b[0:]))
		amb := int16(binary.LittleEndian.Uint16(b[2:]))

		f(float64(obj)/128.0, float64(amb)/128.0)
	})
}

func (c *CC2650) UnsubscribeLuxometer() error {
	return c.client.Unsubscribe(c.chars[LUXOMETER_DATA_UUID], false)
}

// MPU9250

func (c *CC2650) enableMPU9250(mask int) error {
	c.mpu9250mask |= mask

	// for now, always write 0x007f, magnetometer does not seem to notify is specific mask is used
	return c.client.WriteCharacteristic(c.chars[MPU9250_CONFIG_UUID], []byte{0x7f}, false)
}

func convBuf(b []byte, i int, r float64) float64 {
	return float64(int16(binary.LittleEndian.Uint16(b[i:]))) * r
}

func convertMPU9250Data(b []byte) (ret [][]float64) {
	// return will be
	// 250 deg/s range (xG,yG,zG)
	// we specify 8G range in setup (x,y,z)
	// magnetometer (page 50 of http://www.invensense.com/mems/gyro/documents/RM-MPU-9250A-00.pdf) (zM,yM,zM)
	ratio := []float64{1 / 128.0, 1 / 4096.0, 4912.0 / 32768.0}

	for j, r := range ratio {
		vals := make([]float64, 3)
		for i := range vals {
			vals[i] = convBuf(b, j*6+i*2, r)
		}
		ret = append(ret, vals)
	}

	return
}

func (c *CC2650) onMPU9250Change(b []byte) {
	vals := convertMPU9250Data(b)
	if (c.mpu9250mask&MPU9250_ACCELEROMETER_MASK != 0) && c.accelCallback != nil {
		v := vals[1]
		c.accelCallback(v[0], v[1], v[2])
	}
	if (c.mpu9250mask&MPU9250_GYROSCOPE_MASK != 0) && c.gyroCallback != nil {
		v := vals[0]
		c.gyroCallback(v[0], v[1], v[2])
	}
	if (c.mpu9250mask&MPU9250_MAGNETOMETER_MASK != 0) && c.magCallback != nil {
		v := vals[2]
		c.magCallback(v[0], v[1], v[2])
	}
}

func (c *CC2650) subscribeMPU9250() error {
	c.mpu9250Count++

	if c.mpu9250Count == 1 {
		return c.client.Subscribe(c.chars[MPU9250_DATA_UUID], false, func(b []byte) {
			c.onMPU9250Change(b)
		})
	}
	return nil
}

func (c *CC2650) unsubscribeMPU9250() error {
	c.mpu9250Count--

	if c.mpu9250Count == 0 {
		return c.client.Unsubscribe(c.chars[MPU9250_DATA_UUID], false)
	}
	return nil
}

func (c *CC2650) EnableAccelerometer() error {
	return c.enableMPU9250(MPU9250_ACCELEROMETER_MASK)
}

func (c *CC2650) SubscribeAccelerometer(f func(float64, float64, float64)) error {
	c.accelCallback = f
	return c.subscribeMPU9250()
}

func (c *CC2650) UnsubscribeAccelerometer() error {
	c.accelCallback = nil
	return c.unsubscribeMPU9250()
}

func (c *CC2650) EnableGyroscope() error {
	return c.enableMPU9250(MPU9250_GYROSCOPE_MASK)
}

func (c *CC2650) SubscribeGyroscope(f func(float64, float64, float64)) error {
	c.accelCallback = f
	return c.subscribeMPU9250()
}

func (c *CC2650) UnsubscribeGyroscope() error {
	c.accelCallback = nil
	return c.unsubscribeMPU9250()
}

func (c *CC2650) EnableMagnetometer() error {
	return c.enableMPU9250(MPU9250_MAGNETOMETER_MASK)
}

func (c *CC2650) SubscribeMagnetometer(f func(float64, float64, float64)) error {
	c.accelCallback = f
	return c.subscribeMPU9250()
}

func (c *CC2650) UnsubscribeMagnetometer() error {
	c.accelCallback = nil
	return c.unsubscribeMPU9250()
}
