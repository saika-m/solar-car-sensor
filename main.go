package main

import (
	"bufio"
	"fmt"
	"log"
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
	fmt.Println("Received data:", data)

	parts := strings.Split(data, ",")
	fmt.Printf("Number of data points: %d\n", len(parts))

	for i, part := range parts {
		fmt.Printf("Data point %d: %s\n", i+1, part)
	}

	fmt.Println("--------------------")
}
