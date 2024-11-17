package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:9876")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	for {
		fmt.Print("\033[H\033[2J")

		reader := bufio.NewReader(os.Stdin)
		fmt.Println("Enter a command:")
		command, _ := reader.ReadString('\n')

		_, err = conn.Write([]byte(command))
		if err != nil {
			fmt.Println("Error sending message:", err)
			os.Exit(1)
		}
	}
}
