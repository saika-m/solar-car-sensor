package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"go.bug.st/serial"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type SensorData struct {
	Timestamp   int64   `json:"timestamp"`
	Heading     float64 `json:"heading"`
	Roll        float64 `json:"roll"`
	Pitch       float64 `json:"pitch"`
	Temperature float64 `json:"temperature"`
	Pressure    float64 `json:"pressure"`
	Altitude    float64 `json:"altitude"`
	AccX        float64 `json:"accX"`
	AccY        float64 `json:"accY"`
	AccZ        float64 `json:"accZ"`
}

func main() {
	// Open serial port (replace with your Arduino's port)
	port, err := serial.Open("/dev/ttyACM0", &serial.Mode{BaudRate: 115200})
	if err != nil {
		log.Fatal(err)
	}
	defer port.Close()

	http.HandleFunc("/ws", handleConnections)
	go handleSensorData(port)

	fmt.Println("WebSocket server starting on :8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	for {
		// Wait for client to send a message (we don't actually use it)
		_, _, err := ws.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
	}
}

func handleSensorData(port serial.Port) {
	reader := bufio.NewReader(port)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading from serial port: %v", err)
			continue
		}

		data := parseSensorData(strings.TrimSpace(line))
		if data != nil {
			broadcastSensorData(data)
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func parseSensorData(data string) *SensorData {
	parts := strings.Split(data, ",")
	if len(parts) != 9 {
		return nil
	}

	var sensorData SensorData
	var err error

	sensorData.Timestamp = time.Now().UnixNano() / int64(time.Millisecond)
	sensorData.Heading, err = strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return nil
	}
	sensorData.Roll, err = strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return nil
	}
	sensorData.Pitch, err = strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return nil
	}
	sensorData.Temperature, err = strconv.ParseFloat(parts[3], 64)
	if err != nil {
		return nil
	}
	sensorData.Pressure, err = strconv.ParseFloat(parts[4], 64)
	if err != nil {
		return nil
	}
	sensorData.Altitude, err = strconv.ParseFloat(parts[5], 64)
	if err != nil {
		return nil
	}
	sensorData.AccX, err = strconv.ParseFloat(parts[6], 64)
	if err != nil {
		return nil
	}
	sensorData.AccY, err = strconv.ParseFloat(parts[7], 64)
	if err != nil {
		return nil
	}
	sensorData.AccZ, err = strconv.ParseFloat(parts[8], 64)
	if err != nil {
		return nil
	}

	return &sensorData
}

var clients = make(map[*websocket.Conn]bool)

func broadcastSensorData(data *SensorData) {
	json, err := json.Marshal(data)
	if err != nil {
		log.Println("Error marshalling sensor data:", err)
		return
	}

	for client := range clients {
		err := client.WriteMessage(websocket.TextMessage, json)
		if err != nil {
			log.Printf("Error: %v", err)
			client.Close()
			delete(clients, client)
		}
	}
}
