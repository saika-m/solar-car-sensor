package main

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
	fmt.Println("Press Ctrl+C to exit")

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
	validParts := make([]string, 0)

	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			validParts = append(validParts, part)
		}
	}

	if len(validParts) == 0 {
		return // Skip empty lines
	}

	fmt.Println("Valid data points:")
	for i, part := range validParts {
		if value, err := strconv.ParseFloat(part, 64); err == nil {
			fmt.Printf("Data point %d: %.2f\n", i+1, value)
		} else {
			fmt.Printf("Data point %d: %s (non-numeric)\n", i+1, part)
		}
	}

	fmt.Println("--------------------")
}
