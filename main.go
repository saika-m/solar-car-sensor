package main

import (
	"fmt"
	"log"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	BNO055_ADDRESS = 0x28
	BMP280_ADDRESS = 0x76
)

var (
	kernel32        = windows.NewLazySystemDLL("kernel32.dll")
	createFile      = kernel32.NewProc("CreateFileW")
	deviceIoControl = kernel32.NewProc("DeviceIoControl")
)

type I2C_TRANS_DATA struct {
	Address uint16
	Flags   uint16
	Length  uint16
	Buffer  uintptr
}

func main() {
	// Open I2C device
	handle, err := openI2CDevice()
	if err != nil {
		log.Fatal(err)
	}
	defer windows.CloseHandle(handle)

	// Initialize sensors
	if err := initBNO055(handle); err != nil {
		log.Fatal(err)
	}
	if err := initBMP280(handle); err != nil {
		log.Fatal(err)
	}

	for {
		// Read Euler angles from BNO055
		heading, roll, pitch, err := readEulerAngles(handle)
		if err != nil {
			log.Println("Error reading BNO055:", err)
		}

		// Read temperature and pressure from BMP280
		temp, pressure, err := readBMP280(handle)
		if err != nil {
			log.Println("Error reading BMP280:", err)
		}

		fmt.Printf("Heading: %.2f, Roll: %.2f, Pitch: %.2f, Temp: %.2f°C, Pressure: %.2f Pa\n",
			heading, roll, pitch, temp, pressure)

		time.Sleep(100 * time.Millisecond)
	}
}

func openI2CDevice() (windows.Handle, error) {
	deviceName, err := windows.UTF16PtrFromString("\\\\.\\I2C1")
	if err != nil {
		return 0, err
	}
	handle, _, err := createFile.Call(
		uintptr(unsafe.Pointer(deviceName)),
		windows.GENERIC_READ|windows.GENERIC_WRITE,
		0,
		0,
		windows.OPEN_EXISTING,
		0,
		0,
	)
	if handle == windows.InvalidHandle {
		return 0, err
	}
	return windows.Handle(handle), nil
}

func i2cTransfer(handle windows.Handle, addr uint16, writeData []byte, readData []byte) error {
	var bytesReturned uint32
	writeBuffer := I2C_TRANS_DATA{
		Address: addr,
		Flags:   0,
		Length:  uint16(len(writeData)),
		Buffer:  uintptr(unsafe.Pointer(&writeData[0])),
	}
	readBuffer := I2C_TRANS_DATA{
		Address: addr,
		Flags:   1,
		Length:  uint16(len(readData)),
		Buffer:  uintptr(unsafe.Pointer(&readData[0])),
	}
	_, _, err := deviceIoControl.Call(
		uintptr(handle),
		0x22C004, // IOCTL_I2C_TRANSFER
		uintptr(unsafe.Pointer(&writeBuffer)),
		unsafe.Sizeof(writeBuffer),
		uintptr(unsafe.Pointer(&readBuffer)),
		unsafe.Sizeof(readBuffer),
		uintptr(unsafe.Pointer(&bytesReturned)),
		0,
	)
	if err != windows.ERROR_SUCCESS {
		return err
	}
	return nil
}

func initBNO055(handle windows.Handle) error {
	// Set operation mode to NDOF
	return i2cTransfer(handle, BNO055_ADDRESS, []byte{0x3D, 0x0C}, nil)
}

func initBMP280(handle windows.Handle) error {
	// Set oversample for pressure and temperature and set mode to normal
	return i2cTransfer(handle, BMP280_ADDRESS, []byte{0xF4, 0x57}, nil)
}

func readEulerAngles(handle windows.Handle) (float64, float64, float64, error) {
	data := make([]byte, 6)
	if err := i2cTransfer(handle, BNO055_ADDRESS, []byte{0x1A}, data); err != nil {
		return 0, 0, 0, err
	}

	heading := float64(int16(data[0])|int16(data[1])<<8) / 16.0
	roll := float64(int16(data[2])|int16(data[3])<<8) / 16.0
	pitch := float64(int16(data[4])|int16(data[5])<<8) / 16.0

	return heading, roll, pitch, nil
}

func readBMP280(handle windows.Handle) (float64, float64, error) {
	data := make([]byte, 6)
	if err := i2cTransfer(handle, BMP280_ADDRESS, []byte{0xF7}, data); err != nil {
		return 0, 0, err
	}

	pressure := float64((uint32(data[0])<<12)|(uint32(data[1])<<4)|(uint32(data[2])>>4)) / 100.0
	temp := float64((uint32(data[3])<<12)|(uint32(data[4])<<4)|(uint32(data[5])>>4)) / 100.0

	return temp, pressure, nil
}
