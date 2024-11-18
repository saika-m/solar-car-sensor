package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.bug.st/serial"
)

// SensorData struct to hold all the measurements
type SensorData struct {
	// Accelerometer data (mg)
	AccX, AccY, AccZ float64 `json:"accX,accY,accZ"`
	// Magnetometer data (µT)
	MagX, MagY, MagZ float64 `json:"magX,magY,magZ"`
	// Gyroscope data (dps)
	GyrX, GyrY, GyrZ float64 `json:"gyrX,gyrY,gyrZ"`
	// Linear acceleration (mg)
	LiaX, LiaY, LiaZ float64 `json:"liaX,liaY,liaZ"`
	// Gravity vector (mg)
	GrvX, GrvY, GrvZ float64 `json:"grvX,grvY,grvZ"`
	// Euler angles (degrees)
	EulHeading, EulRoll, EulPitch float64 `json:"eulHeading,eulRoll,eulPitch"`
	// Quaternion (no unit)
	QuaW, QuaX, QuaY, QuaZ float64 `json:"quaW,quaX,quaY,quaZ"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var (
	serialPort serial.Port
	portMutex  sync.Mutex
)

func findAndOpenPort() (serial.Port, error) {
	ports, err := serial.GetPortsList()
	if err != nil {
		return nil, fmt.Errorf("error listing ports: %v", err)
	}

	fmt.Println("Available ports:")
	for _, port := range ports {
		fmt.Printf("- %v\n", port)
	}

	// Try to open each port
	for _, portName := range ports {
		fmt.Printf("Trying port %s...\n", portName)
		port, err := serial.Open(portName, &serial.Mode{BaudRate: 115200})
		if err == nil {
			fmt.Printf("Successfully opened port %s\n", portName)
			return port, nil
		}
		fmt.Printf("Failed to open %s: %v\n", portName, err)
	}

	return nil, fmt.Errorf("no available serial ports found")
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Get or create serial port connection
	portMutex.Lock()
	if serialPort == nil {
		var err error
		serialPort, err = findAndOpenPort()
		if err != nil {
			portMutex.Unlock()
			log.Printf("Failed to open any serial port: %v", err)
			return
		}
	}
	portMutex.Unlock()

	reader := bufio.NewReader(serialPort)
	sensorData := &SensorData{}

	// Keep trying to read from the port
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Serial read error: %v", err)
			// Try to reopen port
			portMutex.Lock()
			if serialPort != nil {
				serialPort.Close()
			}
			serialPort, err = findAndOpenPort()
			portMutex.Unlock()
			if err != nil {
				log.Printf("Failed to reopen port: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}
			reader = bufio.NewReader(serialPort)
			continue
		}

		// Process the data
		if processData(strings.TrimSpace(line), sensorData) {
			err = conn.WriteJSON(sensorData)
			if err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func processData(data string, sensorData *SensorData) bool {
	// Skip start and end markers
	if strings.Contains(data, "print start") || strings.Contains(data, "print end") {
		return false
	}

	// Regular expression to extract numbers
	re := regexp.MustCompile(`[-]?\d+\.\d+`)
	updated := false

	// Print raw data for debugging
	fmt.Printf("Raw data: %s\n", data)

	// Parse different sensor data based on the line prefix
	switch {
	case strings.Contains(data, "acc analog"):
		numbers := re.FindAllString(data, -1)
		fmt.Printf("Found numbers for acc: %v\n", numbers)
		if len(numbers) >= 3 {
			fmt.Sscanf(strings.Join(numbers, " "), "%f %f %f",
				&sensorData.AccX, &sensorData.AccY, &sensorData.AccZ)
			updated = true
		}

	case strings.Contains(data, "mag analog"):
		numbers := re.FindAllString(data, -1)
		if len(numbers) >= 3 {
			fmt.Sscanf(strings.Join(numbers, " "), "%f %f %f",
				&sensorData.MagX, &sensorData.MagY, &sensorData.MagZ)
			updated = true
		}

	case strings.Contains(data, "gyr analog"):
		numbers := re.FindAllString(data, -1)
		if len(numbers) >= 3 {
			fmt.Sscanf(strings.Join(numbers, " "), "%f %f %f",
				&sensorData.GyrX, &sensorData.GyrY, &sensorData.GyrZ)
			updated = true
		}

	case strings.Contains(data, "lia analog"):
		numbers := re.FindAllString(data, -1)
		if len(numbers) >= 3 {
			fmt.Sscanf(strings.Join(numbers, " "), "%f %f %f",
				&sensorData.LiaX, &sensorData.LiaY, &sensorData.LiaZ)
			updated = true
		}

	case strings.Contains(data, "grv analog"):
		numbers := re.FindAllString(data, -1)
		if len(numbers) >= 3 {
			fmt.Sscanf(strings.Join(numbers, " "), "%f %f %f",
				&sensorData.GrvX, &sensorData.GrvY, &sensorData.GrvZ)
			updated = true
		}

	case strings.Contains(data, "eul analog"):
		numbers := re.FindAllString(data, -1)
		if len(numbers) >= 3 {
			fmt.Sscanf(strings.Join(numbers, " "), "%f %f %f",
				&sensorData.EulHeading, &sensorData.EulRoll, &sensorData.EulPitch)
			updated = true
		}

	case strings.Contains(data, "qua analog"):
		numbers := re.FindAllString(data, -1)
		if len(numbers) >= 4 {
			fmt.Sscanf(strings.Join(numbers, " "), "%f %f %f %f",
				&sensorData.QuaW, &sensorData.QuaX, &sensorData.QuaY, &sensorData.QuaZ)
			updated = true
		}
	}

	if updated {
		fmt.Printf("Updated sensor data: %+v\n", sensorData)
	}

	return updated
}

func main() {
	// Setup HTTP server
	http.HandleFunc("/ws", handleWebSocket)

	fmt.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
