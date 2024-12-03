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
		HandshakeTimeout: 10 * time.Second,
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
	}

	// Global serial port connection
	serialPort serial.Port

	// Mutex for thread-safe data access
	mutex sync.RWMutex

	// Global sensor data
	currentData = &SensorData{}

	// Active connections
	clients    = make(map[*websocket.Conn]bool)
	clientsMux sync.RWMutex
)

// Initialize serial port connection
func initSerial() error {
	var err error
	mode := &serial.Mode{
		BaudRate: 115200,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
		DataBits: 8,
	}

	serialPort, err = serial.Open("COM3", mode)
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
			time.Sleep(100 * time.Millisecond) // Add delay on error
			continue
		}

		line = strings.TrimSpace(line)

		if strings.Contains(line, "print start") || strings.Contains(line, "print end") {
			continue
		}

		mutex.Lock()
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
			// ... (other cases remain the same)
		}
		mutex.Unlock()

		// Broadcast to all clients
		broadcastSensorData()
	}
}

// Broadcast sensor data to all connected clients
func broadcastSensorData() {
	mutex.RLock()
	data := currentData
	mutex.RUnlock()

	clientsMux.RLock()
	for client := range clients {
		err := client.WriteJSON(data)
		if err != nil {
			log.Printf("Error broadcasting to client: %v", err)
			client.Close()
			clientsMux.RUnlock()
			clientsMux.Lock()
			delete(clients, client)
			clientsMux.Unlock()
			clientsMux.RLock()
		}
	}
	clientsMux.RUnlock()
}

func handleSensor(w http.ResponseWriter, r *http.Request) {
	// Configure WebSocket connection
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Set connection properties
	ws.SetReadLimit(512)
	ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
	ws.SetPongHandler(func(string) error {
		ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
		return nil
	})

	// Add to active clients
	clientsMux.Lock()
	clients[ws] = true
	clientsMux.Unlock()

	// Start ping-pong
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			<-ticker.C
			if err := ws.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second)); err != nil {
				return
			}
		}
	}()

	// Clean up on exit
	defer func() {
		clientsMux.Lock()
		delete(clients, ws)
		clientsMux.Unlock()
		ws.Close()
	}()

	// Keep connection alive
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if err := initSerial(); err != nil {
		log.Fatal(err)
	}
	defer serialPort.Close()

	http.HandleFunc("/sensor", handleSensor)

	server := &http.Server{
		Addr:              ":8080",
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
	}

	fmt.Println("Starting sensor microservice on :8080")
	log.Fatal(server.ListenAndServe())
}
