package ard_main

import (
	"bufio"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"go.bug.st/serial"
)

func main() {
	// List available ports
	ports, err := serial.GetPortsList()
	if err != nil {
		log.Fatal(err)
	}
	if len(ports) == 0 {
		log.Fatal("No serial ports found!")
	}

	fmt.Println("Available ports:")
	for _, port := range ports {
		fmt.Printf("- %v\n", port)
	}

	// Prompt user to select a port
	var selectedPort string
	fmt.Print("Enter the port to use: ")
	fmt.Scanln(&selectedPort)

	// Open the serial port
	mode := &serial.Mode{
		BaudRate: 115200,
	}
	port, err := serial.Open(selectedPort, mode)
	if err != nil {
		log.Fatal(err)
	}
	defer port.Close()

	reader := bufio.NewReader(port)

	fmt.Println("Reading sensor data from Arduino...")

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading from serial port: %v", err)
			continue
		}

		// Process the received data
		processData(strings.TrimSpace(line))

		time.Sleep(100 * time.Millisecond)
	}
}

func processData(data string) {
	parts := strings.Split(data, ",")
	if len(parts) != 9 {
		log.Printf("Unexpected data format: %s", data)
		return
	}

	heading, _ := strconv.ParseFloat(parts[0], 64)
	roll, _ := strconv.ParseFloat(parts[1], 64)
	pitch, _ := strconv.ParseFloat(parts[2], 64)
	temp, _ := strconv.ParseFloat(parts[3], 64)
	pressure, _ := strconv.ParseFloat(parts[4], 64)
	altitude, _ := strconv.ParseFloat(parts[5], 64)
	accX, _ := strconv.ParseFloat(parts[6], 64)
	accY, _ := strconv.ParseFloat(parts[7], 64)
	accZ, _ := strconv.ParseFloat(parts[8], 64)

	fmt.Printf("Heading: %.2f, Roll: %.2f, Pitch: %.2f\n", heading, roll, pitch)
	fmt.Printf("Temperature: %.2f°C, Pressure: %.2f Pa, Altitude: %.2f m\n", temp, pressure, altitude)
	fmt.Printf("Acceleration: X=%.2f, Y=%.2f, Z=%.2f m/s²\n", accX, accY, accZ)
	fmt.Println("--------------------")
}
