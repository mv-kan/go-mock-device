package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"

	gocomm "github.com/mv-kan/go-comm"
)

func waitForFile(filePath string, polling, timeout time.Duration) error {
	start := time.Now()

	// Poll for the file's existence
	for {
		if _, err := os.Stat(filePath); err == nil {
			fmt.Printf("File %s created!\n", filePath)
			return nil
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("error checking file: %v", err)
		}

		// Break if timeout is reached
		if time.Since(start) > timeout {
			return fmt.Errorf("timeout: file %s was not created in time", filePath)
		}

		// Sleep before the next check
		time.Sleep(polling)
	}
}
func main() {
	// log.SetOutput(io.Discard)
	// Check if any arguments are passed
	if len(os.Args) < 2 {
		fmt.Println("No devpath provided. for e.g. pass this path /tmp/virtual_dev")
		return
	}

	// Get the first argument (after the program name)
	// create virtual dev
	devPath := os.Args[1]
	devPathOE := fmt.Sprintf("%s_oe", devPath)
	cmd := exec.Command("socat", "-d", "-d", fmt.Sprintf("pty,raw,echo=0,link=%s", devPath), fmt.Sprintf("pty,raw,echo=0,link=%s", devPathOE))
	cmd.Start()
	waitForFile(devPath, time.Millisecond*50, time.Second*10)
	defer cmd.Process.Signal(syscall.SIGINT)

	baudRate := 115200
	if len(os.Args) >= 3 {
		tmp, err := strconv.Atoi(os.Args[2])
		if err != nil {
			panic(err)
		}
		baudRate = tmp
	}
	port, err := gocomm.NewPort(devPathOE, baudRate, 0)
	if err != nil {
		panic(err)
	}
	input := make(chan string)
	conn, msgChan, err := gocomm.NewConnection(port, input, 0, "\n", "\n")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	go func() {
		for {
			msg := <-msgChan
			fmt.Printf("msg=%+v\n", msg)
		}
	}()
	go func() {
		i := 0
		for {
			input <- fmt.Sprintf("Virtual Device message nr %d", i)
			i++
			time.Sleep(time.Second)
		}
	}()

	select {}
}
