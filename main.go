package main

import (
	"fmt"
	"log"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	BNO055_ADDRESS   = 0x28
	BMP280_ADDRESS   = 0x76
	FILE_SHARE_READ  = 0x00000001
	FILE_SHARE_WRITE = 0x00000002
	OPEN_EXISTING    = 3
	GENERIC_READ     = 0x80000000
	GENERIC_WRITE    = 0x40000000
)

var (
	kernel32        = windows.NewLazySystemDLL("kernel32.dll")
	createFile      = kernel32.NewProc("CreateFileW")
	deviceIoControl = kernel32.NewProc("DeviceIoControl")
)

type I2C_TRANSFER struct {
	DataAddress uint32
	Flags       uint32
	Length      uint32
	Buffer      uintptr
}

func main() {
	handle, err := openI2CDevice()
	if err != nil {
		log.Fatal(err)
	}
	defer windows.CloseHandle(handle)

	for {
		// Read Euler angles from BNO055
		heading, roll, pitch, err := readEulerAngles(handle)
		if err != nil {
			log.Println("Error reading BNO055:", err)
		} else {
			fmt.Printf("Heading: %.2f, Roll: %.2f, Pitch: %.2f\n", heading, roll, pitch)
		}

		// Read temperature and pressure from BMP280
		temp, pressure, err := readBMP280(handle)
		if err != nil {
			log.Println("Error reading BMP280:", err)
		} else {
			fmt.Printf("Temperature: %.2f°C, Pressure: %.2f Pa\n", temp, pressure)
		}

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
		GENERIC_READ|GENERIC_WRITE,
		FILE_SHARE_READ|FILE_SHARE_WRITE,
		0,
		OPEN_EXISTING,
		0,
		0,
	)
	if handle == uintptr(windows.InvalidHandle) {
		return windows.InvalidHandle, err
	}
	return windows.Handle(handle), nil
}

func i2cTransfer(handle windows.Handle, addr uint16, writeData []byte, readData []byte) error {
	var bytesReturned uint32
	writeTransfer := I2C_TRANSFER{
		DataAddress: uint32(addr),
		Flags:       0,
		Length:      uint32(len(writeData)),
		Buffer:      uintptr(unsafe.Pointer(&writeData[0])),
	}
	readTransfer := I2C_TRANSFER{
		DataAddress: uint32(addr),
		Flags:       1,
		Length:      uint32(len(readData)),
		Buffer:      uintptr(unsafe.Pointer(&readData[0])),
	}
	ret, _, err := deviceIoControl.Call(
		uintptr(handle),
		0x22C004, // IOCTL_I2C_TRANSFER
		uintptr(unsafe.Pointer(&writeTransfer)),
		unsafe.Sizeof(writeTransfer),
		uintptr(unsafe.Pointer(&readTransfer)),
		unsafe.Sizeof(readTransfer),
		uintptr(unsafe.Pointer(&bytesReturned)),
		0,
	)
	if ret == 0 {
		return err
	}
	return nil
}

func readEulerAngles(handle windows.Handle) (float32, float32, float32, error) {
	data := make([]byte, 6)
	if err := i2cTransfer(handle, BNO055_ADDRESS, []byte{0x1A}, data); err != nil {
		return 0, 0, 0, err
	}

	heading := float32(int16(data[0])|int16(data[1])<<8) / 16.0
	roll := float32(int16(data[2])|int16(data[3])<<8) / 16.0
	pitch := float32(int16(data[4])|int16(data[5])<<8) / 16.0

	return heading, roll, pitch, nil
}

func readBMP280(handle windows.Handle) (float32, float32, error) {
	data := make([]byte, 6)
	if err := i2cTransfer(handle, BMP280_ADDRESS, []byte{0xF7}, data); err != nil {
		return 0, 0, err
	}

	pressure := float32((uint32(data[0])<<12)|(uint32(data[1])<<4)|(uint32(data[2])>>4)) / 100.0
	temp := float32((uint32(data[3])<<12)|(uint32(data[4])<<4)|(uint32(data[5])>>4)) / 100.0

	return temp, pressure, nil
}
