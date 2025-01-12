package i2c

import (
	"fmt"
	"os"
	"syscall"
)

// I2CDevice represents an I2C device
type I2CDevice struct {
	File *os.File
}

// I2C Constants
const (
	I2C_SLAVE = 0x0703
)

// Init initializes the I2C bus and returns an I2CDevice
func Init(bus int, address uint8) (*I2CDevice, error) {
	filename := fmt.Sprintf("/dev/i2c-%d", bus)
	file, err := os.OpenFile(filename, os.O_RDWR, os.ModeExclusive)
	if err != nil {
		return nil, err
	}

	if err := ioctl(file.Fd(), I2C_SLAVE, uintptr(address)); err != nil {
		return nil, err
	}

	return &I2CDevice{File: file}, nil
}

// Read reads bytes from the I2C device
func (dev *I2CDevice) Read(buf []byte) (int, error) {
	return dev.File.Read(buf)
}

// Write writes bytes to the I2C device
func (dev *I2CDevice) Write(buf []byte) (int, error) {
	return dev.File.Write(buf)
}

// Close closes the I2C device
func (dev *I2CDevice) Close() error {
	return dev.File.Close()
}

// ioctl performs an IO control operation
func ioctl(fd uintptr, request uint, argp uintptr) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(request), argp)
	if errno != 0 {
		return errno
	}
	return nil
}
