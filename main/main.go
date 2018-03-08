package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-ble/ble"
	"github.com/go-ble/ble/examples/lib/dev"
	sensortag "github.com/honteng/go-sensortag"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func startScan() {
	d, err := dev.NewDevice("default")
	check(err)
	ble.SetDefaultDevice(d)

	ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), 100*time.Second))
	cln, err := ble.Connect(ctx, sensortag.Filter)
	check(err)

	fmt.Printf("Discovering profile...\n")
	p, err := cln.DiscoverProfile(true)
	if err != nil {
		log.Fatalf("can't discover profile: %s", err)
	}

	// Start the exploration.
	explore(cln, p)

	//cln.CancelConnection()

	<-cln.Disconnected()
}

func explore(c ble.Client, p *ble.Profile) {
	cc2650, err := sensortag.NewCC2650(c, p)
	check(err)

	cc2650.EnableIrTemperature()

	err = cc2650.SubscribeIrTemperature(func(obj, amb float64) {
		fmt.Printf("%v, %v\n", obj, amb)
	})
	check(err)

	time.Sleep(10 * time.Second)
}

func main() {
	startScan()
}
