package main

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"
)

const (
	START_PORT = 1025
	END_PORT   = 65535
	BATCH_SIZE = 10000
)

var (
	SKIP_PORTS = []int{5357, 5358, 5986}
)

func main() {
	var wg sync.WaitGroup
	success := true
	var counter uint64 = 0

	portCount := END_PORT - START_PORT - len(SKIP_PORTS)
	batchCount := portCount / BATCH_SIZE
	for i := 0; i <= batchCount; i++ {
		startPort := START_PORT + i*BATCH_SIZE
		endPort := min(startPort+BATCH_SIZE, END_PORT)
		wg.Add(1)
		go func(startPort, endPort int) {
			for port := startPort; port < endPort; port++ {
				if contains(SKIP_PORTS, port) {
					continue
				}
				output, err := exec.Command("netsh", "http", "add", "urlacl", fmt.Sprintf("http://*:%d/", port), "user=vcap").CombinedOutput()
				if err != nil {
					fmt.Printf("Failed to add urlacl for port %d: %s\n", port, string(output))
					success = false
				}
				atomic.AddUint64(&counter, 1)
			}
			wg.Done()
		}(startPort, endPort)
	}

	go func() {
		var count uint64 = 0
		for count < uint64(portCount) {
			fmt.Printf("Added urlacl for %d/%d ports\n", count, portCount)
			time.Sleep(time.Minute)
			count = atomic.LoadUint64(&counter)
		}
	}()

	wg.Wait()
	if !success {
		os.Exit(1)
	}
	fmt.Println("Finished adding urlacls")
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func contains(array []int, element int) bool {
	found := false
	for _, e := range array {
		if element == e {
			found = true
			break
		}
	}
	return found
}
