package main

import (
	"bufio"
	"fmt"
	"net"

	"github.com/umran/epigo/model"
)

func main() {
	ln, err := net.Listen("tcp", ":9876")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer ln.Close()

	fmt.Println("Waiting for player to connect on port 9876")

	conn, err := ln.Accept()
	if err != nil {
		fmt.Println("Error accepting connection:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Player connected")
	fmt.Println("Starting game...")

	// a covid-like pathogen
	pathogen := model.Pathogen{
		IncubationPeriod:   [2]float64{3 * 24 * 60 * 60 * 1000, 8 * 60 * 60 * 1000},
		RecoveryPeriod:     [2]float64{7 * 24 * 60 * 60 * 1000, 8 * 60 * 60 * 1000},
		ImmunityPeriod:     [2]float64{330 * 24 * 60 * 60 * 1000, 90 * 24 * 60 * 60 * 1000},
		QuantaEmissionRate: [2]float64{250, 100},
	}

	// create a game with 150k people
	sim := model.NewSimulation(150000, 15*60*1000, pathogen)

	// start a new metrics instance subscribed to simulation events
	sim.Subscribe(model.NewMetricsSubscriber())

	go func() {
		for {
			command, _ := bufio.NewReader(conn).ReadString('\n')
			sim.SendCommand(model.Command(command))
		}
	}()

	sim.Start()
}
