package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type PortInfo struct {
	Port        int
	Banner      string
	ServiceInfo string
}

func main() {
	fmt.Println("Scanning 65,000+ ports. This may take a minute or two. Please wait...")

	// Create a channel to signal when the progress bar should stop
	stopProgressBar := make(chan bool)

	// Start the progress bar in a separate goroutine
	go showProgressBar(stopProgressBar)

	// Get open ports and their information
	openPorts := getOpenPorts()

	// Signal the progress bar to stop
	stopProgressBar <- true

	// Display information about each open port
	for _, portInfo := range openPorts {
		fmt.Printf("Port %d:\nBanner: %s\nService Information: %s\n\n", portInfo.Port, portInfo.Banner, portInfo.ServiceInfo)
	}

	// Prompt user to close a specific open port
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Do you want to close a specific open port? (Y/N): ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if strings.ToLower(input) == "y" {
		// Ask user for the port to close
		fmt.Print("Enter the port number you want to close: ")
		portInput, _ := reader.ReadString('\n')
		portInput = strings.TrimSpace(portInput)
		portToClose, err := strconv.Atoi(portInput)
		if err != nil {
			fmt.Println("Invalid port number. Please enter a valid port number.")
			return
		}

		// Close the specified port
		closePort(portToClose, reader)
	}
}

func showProgressBar(stop chan bool) {
	for i := 0; i < 50; i++ {
		select {
		case <-stop:
			fmt.Println("\nProgress complete.")
			return
		default:
			fmt.Print("\u2588")
			time.Sleep(1200 * time.Millisecond) // Slower progress bar
		}
	}
	fmt.Println()
}

func getOpenPorts() []PortInfo {
	openPorts := []PortInfo{}

	for port := 1; port <= 65535; port++ {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), 2*time.Second)
		if err == nil {
			banner, serviceInfo := getPortDetails(conn)
			openPorts = append(openPorts, PortInfo{Port: port, Banner: banner, ServiceInfo: serviceInfo})
			conn.Close()
		}
	}

	return openPorts
}

func getPortDetails(conn net.Conn) (string, string) {
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))

	// Send specific request to gather more information about the service
	// For demonstration purposes, we simply read some additional information
	// Replace this with actual logic to gather service information

	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err != nil {
		return "Error reading from connection", "No additional information available"
	}

	serviceInfo := string(buffer[:n])
	return "Banner information", serviceInfo
}

func closePort(portToClose int, reader *bufio.Reader) {
	// Check which process is using the port
	cmd := exec.Command("lsof", "-i", fmt.Sprintf(":%d", portToClose))
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error checking process using port %d: %v\n", portToClose, err)
		return
	}
	fmt.Println(string(output))

	// Ask user for confirmation to close the identified program
	fmt.Print("Are you sure you want to close this program? (Y/N): ")
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(confirm)

	if strings.ToLower(confirm) == "y" {
		// Identify and terminate the process using the port
		lines := strings.Split(string(output), "\n")
		if len(lines) < 2 {
			fmt.Println("No process found using port", portToClose)
			return
		}

		fields := strings.Fields(lines[1])
		if len(fields) < 2 {
			fmt.Println("No process found using port", portToClose)
			return
		}

		pid := fields[1]

		// Terminate the identified process
		killCmd := exec.Command("kill", pid)
		if err := killCmd.Run(); err != nil {
			fmt.Println("Error terminating process:", err)
			return
		}

		fmt.Println("Process running on port", portToClose, "has been terminated.")
	} else {
		fmt.Println("Program closure cancelled.")
	}
}
