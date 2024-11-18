package main

import (
	"bufio"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"go.bug.st/serial"
)

// SensorData struct to hold all the measurements
type SensorData struct {
	// Accelerometer data (mg)
	AccX, AccY, AccZ float64
	// Magnetometer data (µT)
	MagX, MagY, MagZ float64
	// Gyroscope data (dps)
	GyrX, GyrY, GyrZ float64
	// Linear acceleration (mg)
	LiaX, LiaY, LiaZ float64
	// Gravity vector (mg)
	GrvX, GrvY, GrvZ float64
	// Euler angles (degrees)
	EulHeading, EulRoll, EulPitch float64
	// Quaternion (no unit)
	QuaW, QuaX, QuaY, QuaZ float64
}

func main() {
	selectedPort := "COM3" // Fixed to COM3

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
	var sensorData SensorData

	fmt.Println("Reading sensor data from COM3...")
	fmt.Println("Press Ctrl+C to exit")

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading from serial port: %v", err)
			continue
		}

		// Process the received data
		processData(strings.TrimSpace(line), &sensorData)

		time.Sleep(100 * time.Millisecond)
	}
}

func processData(data string, sensorData *SensorData) {
	// Skip start and end markers
	if strings.Contains(data, "print start") || strings.Contains(data, "print end") {
		return
	}

	// Regular expression to extract numbers
	re := regexp.MustCompile(`[-]?\d+\.\d+`)

	// Parse different sensor data based on the line prefix
	switch {
	case strings.Contains(data, "acc analog"):
		numbers := re.FindAllString(data, -1)
		if len(numbers) >= 3 {
			fmt.Sscanf(strings.Join(numbers, " "), "%f %f %f",
				&sensorData.AccX, &sensorData.AccY, &sensorData.AccZ)
			printAccelerometer(*sensorData)
		}

	case strings.Contains(data, "mag analog"):
		numbers := re.FindAllString(data, -1)
		if len(numbers) >= 3 {
			fmt.Sscanf(strings.Join(numbers, " "), "%f %f %f",
				&sensorData.MagX, &sensorData.MagY, &sensorData.MagZ)
			printMagnetometer(*sensorData)
		}

	case strings.Contains(data, "gyr analog"):
		numbers := re.FindAllString(data, -1)
		if len(numbers) >= 3 {
			fmt.Sscanf(strings.Join(numbers, " "), "%f %f %f",
				&sensorData.GyrX, &sensorData.GyrY, &sensorData.GyrZ)
			printGyroscope(*sensorData)
		}

	case strings.Contains(data, "lia analog"):
		numbers := re.FindAllString(data, -1)
		if len(numbers) >= 3 {
			fmt.Sscanf(strings.Join(numbers, " "), "%f %f %f",
				&sensorData.LiaX, &sensorData.LiaY, &sensorData.LiaZ)
			printLinearAcceleration(*sensorData)
		}

	case strings.Contains(data, "grv analog"):
		numbers := re.FindAllString(data, -1)
		if len(numbers) >= 3 {
			fmt.Sscanf(strings.Join(numbers, " "), "%f %f %f",
				&sensorData.GrvX, &sensorData.GrvY, &sensorData.GrvZ)
			printGravity(*sensorData)
		}

	case strings.Contains(data, "eul analog"):
		numbers := re.FindAllString(data, -1)
		if len(numbers) >= 3 {
			fmt.Sscanf(strings.Join(numbers, " "), "%f %f %f",
				&sensorData.EulHeading, &sensorData.EulRoll, &sensorData.EulPitch)
			printEuler(*sensorData)
		}

	case strings.Contains(data, "qua analog"):
		numbers := re.FindAllString(data, -1)
		if len(numbers) >= 4 {
			fmt.Sscanf(strings.Join(numbers, " "), "%f %f %f %f",
				&sensorData.QuaW, &sensorData.QuaX, &sensorData.QuaY, &sensorData.QuaZ)
			printQuaternion(*sensorData)
		}
	}
}

// Print functions for each sensor type
func printAccelerometer(data SensorData) {
	fmt.Printf("\nAccelerometer (mg):\n")
	fmt.Printf("  X: %8.2f\n  Y: %8.2f\n  Z: %8.2f\n",
		data.AccX, data.AccY, data.AccZ)
}

func printMagnetometer(data SensorData) {
	fmt.Printf("\nMagnetometer (µT):\n")
	fmt.Printf("  X: %8.2f\n  Y: %8.2f\n  Z: %8.2f\n",
		data.MagX, data.MagY, data.MagZ)
}

func printGyroscope(data SensorData) {
	fmt.Printf("\nGyroscope (dps):\n")
	fmt.Printf("  X: %8.2f\n  Y: %8.2f\n  Z: %8.2f\n",
		data.GyrX, data.GyrY, data.GyrZ)
}

func printLinearAcceleration(data SensorData) {
	fmt.Printf("\nLinear Acceleration (mg):\n")
	fmt.Printf("  X: %8.2f\n  Y: %8.2f\n  Z: %8.2f\n",
		data.LiaX, data.LiaY, data.LiaZ)
}

func printGravity(data SensorData) {
	fmt.Printf("\nGravity Vector (mg):\n")
	fmt.Printf("  X: %8.2f\n  Y: %8.2f\n  Z: %8.2f\n",
		data.GrvX, data.GrvY, data.GrvZ)
}

func printEuler(data SensorData) {
	fmt.Printf("\nEuler Angles (degrees):\n")
	fmt.Printf("  Heading: %8.2f\n  Roll: %8.2f\n  Pitch: %8.2f\n",
		data.EulHeading, data.EulRoll, data.EulPitch)
}

func printQuaternion(data SensorData) {
	fmt.Printf("\nQuaternion:\n")
	fmt.Printf("  W: %8.2f\n  X: %8.2f\n  Y: %8.2f\n  Z: %8.2f\n",
		data.QuaW, data.QuaX, data.QuaY, data.QuaZ)
}
