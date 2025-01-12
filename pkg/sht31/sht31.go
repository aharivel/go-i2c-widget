package sht31

import (
	"math"
	"time"

	"dev/pkg/i2c"
)

const (
	SHT31DefaultAddr        = 0x44   // SHT31 Default Address
	SHT31MeasHighRepStretch = 0x2C06 // Measurement High Repeatability with Clock Stretch Enabled
	SHT31MeasMedRepStretch  = 0x2C0D // Measurement Medium Repeatability with Clock Stretch Enabled
	SHT31MeasLowRepStretch  = 0x2C10 // Measurement Low Repeatability with Clock Stretch Enabled
	SHT31MeasHighRep        = 0x2400 // Measurement High Repeatability with Clock Stretch Disabled
	SHT31MeasMedRep         = 0x240B // Measurement Medium Repeatability with Clock Stretch Disabled
	SHT31MeasLowRep         = 0x2416 // Measurement Low Repeatability with Clock Stretch Disabled
	SHT31ReadStatus         = 0xF32D // Read Out of Status Register
	SHT31ClearStatus        = 0x3041 // Clear Status
	SHT31SoftReset          = 0x30A2 // Soft Reset
	SHT31HeaterEn           = 0x306D // Heater Enable
	SHT31HeaterDis          = 0x3066 // Heater Disable
	SHT31RegHeaterBit       = 0x0D   // Status Register Heater Bit
)

// SHT31Interface defines the methods for interacting with the SHT31 sensor.
type SHT31Interface interface {
	ReadStatus() uint16
	Reset()
	Heater(enable bool)
	IsHeaterEnabled() bool
	ReadTemperature() float64
	ReadHumidity() float64
	ReadBoth() (float64, float64, bool)
}

// SHT31 represents the SHT31 sensor
type SHT31 struct {
	fd       *i2c.I2CDevice
	humidity float64
	temp     float64
}

/////////////////////////////////////////////////////////
//
// # Declaration Functions
//
////////////////////////////////////////////////////////

// SHT31 creates a new instance of the SHT31 sensor
func NewSHT31(fd *i2c.I2CDevice) *SHT31 {
	return &SHT31{
		fd:       fd,
		humidity: math.NaN(),
		temp:     math.NaN(),
	}
}

/////////////////////////////////////////////////////////
//
// # Interface Functions
//
////////////////////////////////////////////////////////

// ReadStatus gets the current status register contents
func (s *SHT31) ReadStatus() uint16 {
	s.WriteCommand(SHT31ReadStatus)

	data := make([]byte, 3)
	s.fd.Read(data)

	stat := uint16(data[0])<<8 | uint16(data[1])
	return stat
}

// Reset performs a reset of the sensor
func (s *SHT31) Reset() {
	s.WriteCommand(SHT31SoftReset)
	time.Sleep(10 * time.Millisecond)
}

// Heater enables or disables the heating element
func (s *SHT31) Heater(enable bool) {
	if enable {
		s.WriteCommand(SHT31HeaterEn)
	} else {
		s.WriteCommand(SHT31HeaterDis)
	}
	time.Sleep(1 * time.Millisecond)
}

// IsHeaterEnabled returns the heater state
func (s *SHT31) IsHeaterEnabled() bool {
	regValue := s.ReadStatus()
	return (regValue>>SHT31RegHeaterBit)&1 == 1
}

// ReadTemperature gets a single temperature reading
func (s *SHT31) ReadTemperature() float64 {
	if !s.ReadTempHum() {
		return math.NaN()
	}
	return s.temp
}

// ReadHumidity gets a single relative humidity reading
func (s *SHT31) ReadHumidity() float64 {
	if !s.ReadTempHum() {
		return math.NaN()
	}
	return s.humidity
}

// ReadBoth gets a reading of both temperature and relative humidity
func (s *SHT31) ReadBoth() (float64, float64, bool) {
	if !s.ReadTempHum() {
		return math.NaN(), math.NaN(), false
	}
	return s.temp, s.humidity, true
}

// ReadTempHum reads temperature and humidity
func (s *SHT31) ReadTempHum() bool {
	readBuffer := make([]byte, 6)

	if _, err := s.WriteCommand(SHT31MeasHighRep); err != nil {
		// Handle the error (e.g., log it, return it, etc.)
		return false
	}

	time.Sleep(20 * time.Millisecond)

	if n, err := s.fd.Read(readBuffer); err != nil || n != len(readBuffer) {
		// Handle the error or the case where not all bytes were read
		return false
	}

	if readBuffer[2] != crc8(readBuffer[:2]) || readBuffer[5] != crc8(readBuffer[3:5]) {
		return false
	}

	stemp := int32(uint32(readBuffer[0])<<8 | uint32(readBuffer[1]))
	stemp = ((4375 * stemp) >> 14) - 4500
	s.temp = float64(stemp) / 100.0

	shum := uint32(readBuffer[3])<<8 | uint32(readBuffer[4])
	shum = (625 * shum) >> 12
	s.humidity = float64(shum) / 100.0

	return true
}

////////////////////////////////////////////////////////
//
// # Private Functions
//
////////////////////////////////////////////////////////

// WriteCommand performs an I2C write with the given command
func (s *SHT31) WriteCommand(command uint16) (int, error) {
	cmd := []byte{byte(command >> 8), byte(command & 0xFF)}
	return s.fd.Write(cmd)
}

// crc8 performs a CRC8 calculation on the supplied values
func crc8(data []byte) uint8 {
	const polynomial = 0x31
	var crc uint8 = 0xFF

	for _, b := range data {
		crc ^= b
		for i := 0; i < 8; i++ {
			if (crc & 0x80) != 0 {
				crc = (crc << 1) ^ polynomial
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}
