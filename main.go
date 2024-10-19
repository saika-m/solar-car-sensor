package main

import (
	"fmt"
	"log"
	"time"

	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

const (
	BNO055_ADDRESS = 0x28
	BMP280_ADDRESS = 0x76
)

func main() {
	// Initialize periph.io
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Open I2C bus
	bus, err := i2creg.Open("")
	if err != nil {
		log.Fatal(err)
	}
	defer bus.Close()

	// Create device for BNO055
	bno055 := &i2c.Dev{Addr: BNO055_ADDRESS, Bus: bus}

	// Create device for BMP280
	bmp280 := &i2c.Dev{Addr: BMP280_ADDRESS, Bus: bus}

	// Initialize sensors
	if err := initBNO055(bno055); err != nil {
		log.Fatal(err)
	}
	if err := initBMP280(bmp280); err != nil {
		log.Fatal(err)
	}

	for {
		// Read Euler angles from BNO055
		heading, roll, pitch, err := readEulerAngles(bno055)
		if err != nil {
			log.Println("Error reading BNO055:", err)
		}

		// Read temperature and pressure from BMP280
		temp, pressure, err := readBMP280(bmp280)
		if err != nil {
			log.Println("Error reading BMP280:", err)
		}

		fmt.Printf("Heading: %.2f, Roll: %.2f, Pitch: %.2f, Temp: %.2f°C, Pressure: %.2f Pa\n",
			heading, roll, pitch, temp, pressure)

		time.Sleep(100 * time.Millisecond)
	}
}

func initBNO055(dev *i2c.Dev) error {
	// Set operation mode to NDOF
	if err := dev.Tx([]byte{0x3D, 0x0C}, nil); err != nil {
		return err
	}
	time.Sleep(20 * time.Millisecond)
	return nil
}

func initBMP280(dev *i2c.Dev) error {
	if err := dev.Tx([]byte{0xF4, 0x57}, nil); err != nil {
		return err
	}
	time.Sleep(20 * time.Millisecond)
	return nil
}

func readEulerAngles(dev *i2c.Dev) (float64, float64, float64, error) {
	data := make([]byte, 6)
	if err := dev.Tx([]byte{0x1A}, data); err != nil {
		return 0, 0, 0, err
	}

	heading := float64(int16(data[0])|int16(data[1])<<8) / 16.0
	roll := float64(int16(data[2])|int16(data[3])<<8) / 16.0
	pitch := float64(int16(data[4])|int16(data[5])<<8) / 16.0

	return heading, roll, pitch, nil
}

func readBMP280(dev *i2c.Dev) (float64, float64, error) {
	data := make([]byte, 6)
	if err := dev.Tx([]byte{0xF7}, data); err != nil {
		return 0, 0, err
	}

	pressure := float64((uint32(data[0])<<12)|(uint32(data[1])<<4)|(uint32(data[2])>>4)) / 100.0
	temp := float64((uint32(data[3])<<12)|(uint32(data[4])<<4)|(uint32(data[5])>>4)) / 100.0

	return temp, pressure, nil
}
