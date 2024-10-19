package main

import (
	"fmt"
	"log"
	"time"

	"github.com/sigurn/crc16"
	"github.com/tarm/serial"
)

func main() {
	c := &serial.Config{Name: "COM3", Baud: 115200}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}

	for {
		// Read Euler angles
		heading, roll, pitch, err := readEulerAngles(s)
		if err != nil {
			log.Println("Error reading Euler angles:", err)
		} else {
			fmt.Printf("Heading: %.2f, Roll: %.2f, Pitch: %.2f\n", heading, roll, pitch)
		}

		// Read temperature and pressure
		temp, pressure, err := readTempPressure(s)
		if err != nil {
			log.Println("Error reading temperature and pressure:", err)
		} else {
			fmt.Printf("Temperature: %.2f°C, Pressure: %.2f Pa\n", temp, pressure)
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func readEulerAngles(s *serial.Port) (float32, float32, float32, error) {
	cmd := []byte{0xA5, 0x45, 0xAA, 0x01, 0x00, 0x00}
	table := crc16.MakeTable(crc16.CRC16_MODBUS)
	crc := crc16.Checksum(cmd[:4], table)
	cmd[4] = byte(crc & 0xFF)
	cmd[5] = byte((crc >> 8) & 0xFF)

	_, err := s.Write(cmd)
	if err != nil {
		return 0, 0, 0, err
	}

	buf := make([]byte, 11)
	_, err = s.Read(buf)
	if err != nil {
		return 0, 0, 0, err
	}

	heading := float32(int16(buf[3])<<8|int16(buf[4])) / 100.0
	roll := float32(int16(buf[5])<<8|int16(buf[6])) / 100.0
	pitch := float32(int16(buf[7])<<8|int16(buf[8])) / 100.0

	return heading, roll, pitch, nil
}

func readTempPressure(s *serial.Port) (float32, float32, error) {
	cmd := []byte{0xA5, 0x45, 0xAA, 0x02, 0x00, 0x00}
	table := crc16.MakeTable(crc16.CRC16_MODBUS)
	crc := crc16.Checksum(cmd[:4], table)
	cmd[4] = byte(crc & 0xFF)
	cmd[5] = byte((crc >> 8) & 0xFF)

	_, err := s.Write(cmd)
	if err != nil {
		return 0, 0, err
	}

	buf := make([]byte, 11)
	_, err = s.Read(buf)
	if err != nil {
		return 0, 0, err
	}

	temp := float32(int16(buf[3])<<8|int16(buf[4])) / 100.0
	pressure := float32(int32(buf[5])<<16|int32(buf[6])<<8|int32(buf[7])) / 100.0

	return temp, pressure, nil
}
