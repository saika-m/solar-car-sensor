package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/gorilla/websocket"
	"go.bug.st/serial"
)

type SensorData struct {
	AccX, AccY, AccZ              float64 `json:"accX,accY,accZ"`
	MagX, MagY, MagZ              float64 `json:"magX,magY,magZ"`
	GyrX, GyrY, GyrZ              float64 `json:"gyrX,gyrY,gyrZ"`
	LiaX, LiaY, LiaZ              float64 `json:"liaX,liaY,liaZ"`
	GrvX, GrvY, GrvZ              float64 `json:"grvX,grvY,grvZ"`
	EulHeading, EulRoll, EulPitch float64 `json:"eulHeading,eulRoll,eulPitch"`
	QuaW, QuaX, QuaY, QuaZ        float64 `json:"quaW,quaX,quaY,quaZ"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleSensor(w http.ResponseWriter, r *http.Request) {
	// Open COM3
	port, err := serial.Open("COM3", &serial.Mode{BaudRate: 115200})
	if err != nil {
		http.Error(w, "Could not open COM3", http.StatusInternalServerError)
		return
	}
	defer port.Close()

	// Upgrade to websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer ws.Close()

	reader := bufio.NewReader(port)
	data := &SensorData{}

	// Regular expression for number extraction
	re := regexp.MustCompile(`[-]?\d+\.\d+`)

	for {
		// Read a line from COM3
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		line = strings.TrimSpace(line)

		// Skip markers
		if strings.Contains(line, "print start") || strings.Contains(line, "print end") {
			continue
		}

		// Extract values based on line prefix
		switch {
		case strings.Contains(line, "acc analog"):
			if nums := re.FindAllString(line, -1); len(nums) >= 3 {
				fmt.Sscanf(strings.Join(nums, " "), "%f %f %f", &data.AccX, &data.AccY, &data.AccZ)
			}
		case strings.Contains(line, "mag analog"):
			if nums := re.FindAllString(line, -1); len(nums) >= 3 {
				fmt.Sscanf(strings.Join(nums, " "), "%f %f %f", &data.MagX, &data.MagY, &data.MagZ)
			}
		case strings.Contains(line, "gyr analog"):
			if nums := re.FindAllString(line, -1); len(nums) >= 3 {
				fmt.Sscanf(strings.Join(nums, " "), "%f %f %f", &data.GyrX, &data.GyrY, &data.GyrZ)
			}
		case strings.Contains(line, "lia analog"):
			if nums := re.FindAllString(line, -1); len(nums) >= 3 {
				fmt.Sscanf(strings.Join(nums, " "), "%f %f %f", &data.LiaX, &data.LiaY, &data.LiaZ)
			}
		case strings.Contains(line, "grv analog"):
			if nums := re.FindAllString(line, -1); len(nums) >= 3 {
				fmt.Sscanf(strings.Join(nums, " "), "%f %f %f", &data.GrvX, &data.GrvY, &data.GrvZ)
			}
		case strings.Contains(line, "eul analog"):
			if nums := re.FindAllString(line, -1); len(nums) >= 3 {
				fmt.Sscanf(strings.Join(nums, " "), "%f %f %f", &data.EulHeading, &data.EulRoll, &data.EulPitch)
			}
		case strings.Contains(line, "qua analog"):
			if nums := re.FindAllString(line, -1); len(nums) >= 4 {
				fmt.Sscanf(strings.Join(nums, " "), "%f %f %f %f", &data.QuaW, &data.QuaX, &data.QuaY, &data.QuaZ)
			}
		}

		// Send the data over websocket
		if err := ws.WriteJSON(data); err != nil {
			break
		}
	}
}

func main() {
	http.HandleFunc("/sensor", handleSensor)
	fmt.Println("Starting sensor microservice on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
