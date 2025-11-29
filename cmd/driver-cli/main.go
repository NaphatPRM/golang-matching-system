package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultServerURL = "http://localhost:8080/api/v1"
)

var (
	serverURL string
	driverID  string
	client    = &http.Client{Timeout: 10 * time.Second}
)

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Driver struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Location     Location `json:"location"`
	Status       string   `json:"status"`
	ActiveOrders []string `json:"active_orders"`
}

type Order struct {
	ID                 string   `json:"id"`
	Status             string   `json:"status"`
	RestaurantID       string   `json:"restaurant_id"`
	RestaurantLocation Location `json:"restaurant_location"`
	DeliveryLocation   Location `json:"delivery_location"`
	TotalAmount        float64  `json:"total_amount"`
}

func main() {
	rand.Seed(time.Now().UnixNano())

	serverURL = os.Getenv("SERVER_URL")
	if serverURL == "" {
		serverURL = defaultServerURL
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("===================================")
	fmt.Println("   Food Delivery Driver CLI App   ")
	fmt.Println("===================================")
	fmt.Printf("Server: %s\n\n", serverURL)

	for {
		if driverID == "" {
			showMainMenu()
		} else {
			showDriverMenu()
		}

		fmt.Print("\nEnter choice: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if driverID == "" {
			handleMainMenu(input, reader)
		} else {
			handleDriverMenu(input, reader)
		}
	}
}

func showMainMenu() {
	fmt.Println("\n--- Main Menu ---")
	fmt.Println("1. Register as new driver")
	fmt.Println("2. Login as existing driver")
	fmt.Println("3. List all drivers")
	fmt.Println("4. Exit")
}

func showDriverMenu() {
	fmt.Printf("\n--- Driver Menu (ID: %s) ---\n", driverID[:8])
	fmt.Println("1. View my info")
	fmt.Println("2. Update my location")
	fmt.Println("3. Go online (available)")
	fmt.Println("4. Go offline")
	fmt.Println("5. View my orders")
	fmt.Println("6. Update order status")
	fmt.Println("7. Simulate location movement")
	fmt.Println("8. View all pending orders")
	fmt.Println("9. Logout")
	fmt.Println("0. Exit")
}

func handleMainMenu(choice string, reader *bufio.Reader) {
	switch choice {
	case "1":
		registerDriver(reader)
	case "2":
		loginDriver(reader)
	case "3":
		listDrivers()
	case "4":
		fmt.Println("Goodbye!")
		os.Exit(0)
	default:
		fmt.Println("Invalid choice")
	}
}

func handleDriverMenu(choice string, reader *bufio.Reader) {
	switch choice {
	case "1":
		viewDriverInfo()
	case "2":
		updateLocation(reader)
	case "3":
		goOnline()
	case "4":
		goOffline()
	case "5":
		viewMyOrders()
	case "6":
		updateOrderStatus(reader)
	case "7":
		simulateMovement(reader)
	case "8":
		viewPendingOrders()
	case "9":
		driverID = ""
		fmt.Println("Logged out successfully")
	case "0":
		fmt.Println("Goodbye!")
		os.Exit(0)
	default:
		fmt.Println("Invalid choice")
	}
}

func registerDriver(reader *bufio.Reader) {
	fmt.Print("Enter your name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	if name == "" {
		fmt.Println("Name cannot be empty")
		return
	}

	body := map[string]interface{}{
		"name":           name,
		"max_concurrent": 3,
	}

	resp, err := doRequest("POST", "/drivers", body)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	var driver Driver
	if err := json.Unmarshal(resp, &driver); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return
	}

	driverID = driver.ID
	fmt.Printf("\n✓ Registered successfully!\n")
	fmt.Printf("  Driver ID: %s\n", driver.ID)
	fmt.Printf("  Name: %s\n", driver.Name)
	fmt.Println("\nNote: You are currently offline. Go online to start receiving orders.")
}

func loginDriver(reader *bufio.Reader) {
	fmt.Print("Enter driver ID: ")
	id, _ := reader.ReadString('\n')
	id = strings.TrimSpace(id)

	if id == "" {
		fmt.Println("ID cannot be empty")
		return
	}

	resp, err := doRequest("GET", "/drivers/"+id, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	var driver Driver
	if err := json.Unmarshal(resp, &driver); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return
	}

	driverID = driver.ID
	fmt.Printf("\n✓ Logged in as %s\n", driver.Name)
	fmt.Printf("  Status: %s\n", driver.Status)
	fmt.Printf("  Active orders: %d\n", len(driver.ActiveOrders))
}

func listDrivers() {
	resp, err := doRequest("GET", "/drivers", nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	var drivers []Driver
	if err := json.Unmarshal(resp, &drivers); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return
	}

	fmt.Printf("\n--- All Drivers (%d) ---\n", len(drivers))
	for _, d := range drivers {
		fmt.Printf("  ID: %s\n", d.ID)
		fmt.Printf("  Name: %s | Status: %s | Orders: %d\n", d.Name, d.Status, len(d.ActiveOrders))
		fmt.Printf("  Location: (%.4f, %.4f)\n", d.Location.Latitude, d.Location.Longitude)
		fmt.Println()
	}
}

func viewDriverInfo() {
	resp, err := doRequest("GET", "/drivers/"+driverID, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	var driver Driver
	if err := json.Unmarshal(resp, &driver); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return
	}

	fmt.Printf("\n--- Driver Info ---\n")
	fmt.Printf("  ID: %s\n", driver.ID)
	fmt.Printf("  Name: %s\n", driver.Name)
	fmt.Printf("  Status: %s\n", driver.Status)
	fmt.Printf("  Location: (%.6f, %.6f)\n", driver.Location.Latitude, driver.Location.Longitude)
	fmt.Printf("  Active orders: %d\n", len(driver.ActiveOrders))
	if len(driver.ActiveOrders) > 0 {
		fmt.Printf("  Order IDs: %v\n", driver.ActiveOrders)
	}
}

func updateLocation(reader *bufio.Reader) {
	fmt.Print("Enter latitude (or press Enter for random): ")
	latStr, _ := reader.ReadString('\n')
	latStr = strings.TrimSpace(latStr)

	var lat, lon float64

	if latStr == "" {
		// Generate random location (somewhere around a typical city center)
		lat = 13.7 + (rand.Float64() * 0.1) // Bangkok area as example
		lon = 100.5 + (rand.Float64() * 0.1)
	} else {
		var err error
		lat, err = strconv.ParseFloat(latStr, 64)
		if err != nil {
			fmt.Println("Invalid latitude")
			return
		}

		fmt.Print("Enter longitude: ")
		lonStr, _ := reader.ReadString('\n')
		lonStr = strings.TrimSpace(lonStr)

		lon, err = strconv.ParseFloat(lonStr, 64)
		if err != nil {
			fmt.Println("Invalid longitude")
			return
		}
	}

	body := map[string]interface{}{
		"location": map[string]float64{
			"latitude":  lat,
			"longitude": lon,
		},
	}

	resp, err := doRequest("PUT", "/drivers/"+driverID+"/location", body)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	var driver Driver
	if err := json.Unmarshal(resp, &driver); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return
	}

	fmt.Printf("\n✓ Location updated to (%.6f, %.6f)\n", driver.Location.Latitude, driver.Location.Longitude)
}

func goOnline() {
	body := map[string]interface{}{
		"status": "available",
	}

	resp, err := doRequest("PUT", "/drivers/"+driverID+"/status", body)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	var driver Driver
	if err := json.Unmarshal(resp, &driver); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return
	}

	fmt.Printf("\n✓ You are now ONLINE and available for orders!\n")
}

func goOffline() {
	body := map[string]interface{}{
		"status": "offline",
	}

	resp, err := doRequest("PUT", "/drivers/"+driverID+"/status", body)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	var driver Driver
	if err := json.Unmarshal(resp, &driver); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return
	}

	fmt.Printf("\n✓ You are now OFFLINE\n")
}

func viewMyOrders() {
	resp, err := doRequest("GET", "/orders/driver/"+driverID, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	var orders []Order
	if err := json.Unmarshal(resp, &orders); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return
	}

	if len(orders) == 0 {
		fmt.Println("\nNo orders assigned to you")
		return
	}

	fmt.Printf("\n--- Your Orders (%d) ---\n", len(orders))
	for i, o := range orders {
		fmt.Printf("\n  %d. Order ID: %s\n", i+1, o.ID)
		fmt.Printf("     Status: %s\n", o.Status)
		fmt.Printf("     Total: $%.2f\n", o.TotalAmount)
		fmt.Printf("     Restaurant: (%.4f, %.4f)\n", o.RestaurantLocation.Latitude, o.RestaurantLocation.Longitude)
		fmt.Printf("     Delivery: (%.4f, %.4f)\n", o.DeliveryLocation.Latitude, o.DeliveryLocation.Longitude)
	}
}

func updateOrderStatus(reader *bufio.Reader) {
	fmt.Print("Enter order ID: ")
	orderID, _ := reader.ReadString('\n')
	orderID = strings.TrimSpace(orderID)

	if orderID == "" {
		fmt.Println("Order ID cannot be empty")
		return
	}

	fmt.Println("Select new status:")
	fmt.Println("  1. picked_up (food collected from restaurant)")
	fmt.Println("  2. delivering (on the way to customer)")
	fmt.Println("  3. delivered (order completed)")
	fmt.Println("  4. cancelled")
	fmt.Print("Choice: ")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	var status string
	switch choice {
	case "1":
		status = "picked_up"
	case "2":
		status = "delivering"
	case "3":
		status = "delivered"
	case "4":
		status = "cancelled"
	default:
		fmt.Println("Invalid choice")
		return
	}

	body := map[string]interface{}{
		"status": status,
	}

	resp, err := doRequest("PUT", "/orders/"+orderID+"/status", body)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	var order Order
	if err := json.Unmarshal(resp, &order); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return
	}

	fmt.Printf("\n✓ Order status updated to: %s\n", order.Status)
}

func simulateMovement(reader *bufio.Reader) {
	fmt.Println("\nThis will simulate real-time location updates.")
	fmt.Print("Enter duration in seconds (default: 30): ")
	durStr, _ := reader.ReadString('\n')
	durStr = strings.TrimSpace(durStr)

	duration := 30
	if durStr != "" {
		if d, err := strconv.Atoi(durStr); err == nil {
			duration = d
		}
	}

	// Get current location
	resp, err := doRequest("GET", "/drivers/"+driverID, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	var driver Driver
	if err := json.Unmarshal(resp, &driver); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return
	}

	lat := driver.Location.Latitude
	lon := driver.Location.Longitude

	// If location is zero, start from a default location
	if lat == 0 && lon == 0 {
		lat = 13.7563 // Bangkok
		lon = 100.5018
	}

	fmt.Printf("\nStarting location simulation for %d seconds...\n", duration)
	fmt.Println("Press Ctrl+C to stop early")

	// Generate a random direction
	angle := rand.Float64() * 2 * math.Pi
	speed := 0.0003 // Roughly 30 m/s or about 100 km/h

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	endTime := time.Now().Add(time.Duration(duration) * time.Second)
	updateCount := 0

	for time.Now().Before(endTime) {
		select {
		case <-ticker.C:
			// Move in the direction with some randomness
			angle += (rand.Float64() - 0.5) * 0.5
			lat += speed * math.Sin(angle)
			lon += speed * math.Cos(angle)

			body := map[string]interface{}{
				"location": map[string]float64{
					"latitude":  lat,
					"longitude": lon,
				},
			}

			_, err := doRequest("PUT", "/drivers/"+driverID+"/location", body)
			if err != nil {
				fmt.Printf("Error updating location: %v\n", err)
				continue
			}

			updateCount++
			fmt.Printf("  [%d] Location: (%.6f, %.6f)\n", updateCount, lat, lon)
		}
	}

	fmt.Printf("\n✓ Simulation complete. Sent %d location updates.\n", updateCount)
}

func viewPendingOrders() {
	resp, err := doRequest("GET", "/orders/pending", nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	var orders []Order
	if err := json.Unmarshal(resp, &orders); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return
	}

	if len(orders) == 0 {
		fmt.Println("\nNo pending orders")
		return
	}

	fmt.Printf("\n--- Pending Orders (%d) ---\n", len(orders))
	for i, o := range orders {
		fmt.Printf("\n  %d. Order ID: %s\n", i+1, o.ID)
		fmt.Printf("     Total: $%.2f\n", o.TotalAmount)
		fmt.Printf("     Restaurant: (%.4f, %.4f)\n", o.RestaurantLocation.Latitude, o.RestaurantLocation.Longitude)
		fmt.Printf("     Delivery: (%.4f, %.4f)\n", o.DeliveryLocation.Latitude, o.DeliveryLocation.Longitude)
	}
}

func doRequest(method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, serverURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("server error (%d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
