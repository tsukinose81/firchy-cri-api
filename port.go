package criapi

import (
	"fmt"
	"math/rand"
	"time"
)

// FindAvailablePort finds an available port in the given range
// Note: This generates a random port number. The actual availability
// is checked when the Pod is created on the remote host.
func FindAvailablePort(minPort, maxPort int32) (int32, error) {
	if minPort >= maxPort {
		return 0, fmt.Errorf("invalid port range: %d-%d", minPort, maxPort)
	}
	
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())
	
	// Generate random port in range
	portRange := maxPort - minPort + 1
	port := minPort + rand.Int31n(portRange)
	
	return port, nil
}

// FindAvailablePortWithRetry tries to find an available port with retry logic
func FindAvailablePortWithRetry(minPort, maxPort int32, maxRetries int) (int32, error) {
	usedPorts := make(map[int32]bool)
	
	for i := 0; i < maxRetries; i++ {
		port, err := FindAvailablePort(minPort, maxPort)
		if err != nil {
			return 0, err
		}
		
		// Skip if we've already tried this port
		if usedPorts[port] {
			continue
		}
		
		usedPorts[port] = true
		return port, nil
	}
	
	return 0, fmt.Errorf("failed to find available port after %d retries", maxRetries)
}
