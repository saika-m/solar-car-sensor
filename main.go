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
	deviceNames := []string{"\\\\.\\4DC5", "\\\\.\\4DC6", "\\\\.\\4DE8", "\\\\.\\4DE9", "\\\\.\\4DEA", "\\\\.\\4DAB"}

	for _, deviceName := range deviceNames {
		handle, err := openI2CDevice(deviceName)
		if err != nil {
			fmt.Printf("Failed to open %s: %v\n", deviceName, err)
			continue
		}
		defer windows.CloseHandle(handle)

		fmt.Printf("Scanning %s...\n", deviceName)
		scanI2C(handle)
		fmt.Println("Scan complete.")

		if devicesFound(handle) {
			fmt.Printf("Devices found on %s\n", deviceName)
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
	}
}

func openI2CDevice(deviceName string) (windows.Handle, error) {
	name, err := windows.UTF16PtrFromString(deviceName)
	if err != nil {
		return 0, err
	}
	handle, _, err := createFile.Call(
		uintptr(unsafe.Pointer(name)),
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

func scanI2C(handle windows.Handle) {
	for addr := uint16(0); addr < 128; addr++ {
		err := i2cTransfer(handle, addr, []byte{0}, []byte{0})
		if err == nil {
			fmt.Printf("Device found at address: 0x%02X\n", addr)
		}
	}
}

func devicesFound(handle windows.Handle) bool {
	err1 := i2cTransfer(handle, BNO055_ADDRESS, []byte{0}, []byte{0})
	err2 := i2cTransfer(handle, BMP280_ADDRESS, []byte{0}, []byte{0})
	return err1 == nil && err2 == nil
}
