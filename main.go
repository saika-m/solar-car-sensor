package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"go.bug.st/serial"
)

type SensorData struct {
	AccX       float64 `json:"accX"`
	AccY       float64 `json:"accY"`
	AccZ       float64 `json:"accZ"`
	MagX       float64 `json:"magX"`
	MagY       float64 `json:"magY"`
	MagZ       float64 `json:"magZ"`
	GyrX       float64 `json:"gyrX"`
	GyrY       float64 `json:"gyrY"`
	GyrZ       float64 `json:"gyrZ"`
	LiaX       float64 `json:"liaX"`
	LiaY       float64 `json:"liaY"`
	LiaZ       float64 `json:"liaZ"`
	GrvX       float64 `json:"grvX"`
	GrvY       float64 `json:"grvY"`
	GrvZ       float64 `json:"grvZ"`
	EulHeading float64 `json:"eulHeading"`
	EulRoll    float64 `json:"eulRoll"`
	EulPitch   float64 `json:"eulPitch"`
	QuaW       float64 `json:"quaW"`
	QuaX       float64 `json:"quaX"`
	QuaY       float64 `json:"quaY"`
	QuaZ       float64 `json:"quaZ"`
}

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	// Global serial port connection
	serialPort serial.Port

	// Mutex for thread-safe data access
	mutex sync.RWMutex

	// Global sensor data
	currentData = &SensorData{}
)

// Initialize serial port connection
func initSerial() error {
	var err error
	serialPort, err = serial.Open("COM3", &serial.Mode{BaudRate: 115200})
	if err != nil {
		return fmt.Errorf("failed to open COM3: %v", err)
	}

	go readSerialData()
	return nil
}

// Continuously read from serial port
func readSerialData() {
	reader := bufio.NewReader(serialPort)
	re := regexp.MustCompile(`[-]?\d+\.\d+`)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading serial: %v", err)
			continue
		}

		line = strings.TrimSpace(line)

		// Skip markers
		if strings.Contains(line, "print start") || strings.Contains(line, "print end") {
			continue
		}

		mutex.Lock()
		// Extract values based on line prefix
		switch {
		case strings.Contains(line, "acc analog"):
			if nums := re.FindAllString(line, -1); len(nums) >= 3 {
				fmt.Sscanf(strings.Join(nums, " "), "%f %f %f",
					&currentData.AccX, &currentData.AccY, &currentData.AccZ)
			}
		case strings.Contains(line, "mag analog"):
			if nums := re.FindAllString(line, -1); len(nums) >= 3 {
				fmt.Sscanf(strings.Join(nums, " "), "%f %f %f",
					&currentData.MagX, &currentData.MagY, &currentData.MagZ)
			}
		case strings.Contains(line, "gyr analog"):
			if nums := re.FindAllString(line, -1); len(nums) >= 3 {
				fmt.Sscanf(strings.Join(nums, " "), "%f %f %f",
					&currentData.GyrX, &currentData.GyrY, &currentData.GyrZ)
			}
		case strings.Contains(line, "lia analog"):
			if nums := re.FindAllString(line, -1); len(nums) >= 3 {
				fmt.Sscanf(strings.Join(nums, " "), "%f %f %f",
					&currentData.LiaX, &currentData.LiaY, &currentData.LiaZ)
			}
		case strings.Contains(line, "grv analog"):
			if nums := re.FindAllString(line, -1); len(nums) >= 3 {
				fmt.Sscanf(strings.Join(nums, " "), "%f %f %f",
					&currentData.GrvX, &currentData.GrvY, &currentData.GrvZ)
			}
		case strings.Contains(line, "eul analog"):
			if nums := re.FindAllString(line, -1); len(nums) >= 3 {
				fmt.Sscanf(strings.Join(nums, " "), "%f %f %f",
					&currentData.EulHeading, &currentData.EulRoll, &currentData.EulPitch)
			}
		case strings.Contains(line, "qua analog"):
			if nums := re.FindAllString(line, -1); len(nums) >= 4 {
				fmt.Sscanf(strings.Join(nums, " "), "%f %f %f %f",
					&currentData.QuaW, &currentData.QuaX, &currentData.QuaY, &currentData.QuaZ)
			}
		}
		mutex.Unlock()
	}
}

func handleSensor(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer ws.Close()

	// Send data periodically
	for {
		mutex.RLock()
		err := ws.WriteJSON(currentData)
		mutex.RUnlock()

		if err != nil {
			log.Printf("WebSocket write error: %v", err)
			break
		}
	}
}

func main() {
	// Initialize serial port
	if err := initSerial(); err != nil {
		log.Fatal(err)
	}
	defer serialPort.Close()

	http.HandleFunc("/sensor", handleSensor)
	fmt.Println("Starting sensor microservice on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
