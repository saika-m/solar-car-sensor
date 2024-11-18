package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
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
		return true // Allow all origins for development
	},
}

func main() {
	// Setup websocket handler
	http.HandleFunc("/ws", handleWebSocket)

	fmt.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Open serial port
	port, err := serial.Open("COM3", &serial.Mode{BaudRate: 115200})
	if err != nil {
		log.Printf("Serial port error: %v", err)
		return
	}
	defer port.Close()

	reader := bufio.NewReader(port)
	sensorData := &SensorData{}

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Serial read error: %v", err)
			continue
		}

		// Process the data
		if processData(strings.TrimSpace(line), sensorData) {
			// Send the updated sensor data through WebSocket
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

	// Parse different sensor data based on the line prefix
	switch {
	case strings.Contains(data, "acc analog"):
		numbers := re.FindAllString(data, -1)
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

	return updated
}
