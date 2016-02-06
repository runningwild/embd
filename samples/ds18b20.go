// +build ignore

package main

import (
	"fmt"
	"strings"

	"github.com/zlowred/embd"
	"github.com/zlowred/embd/sensor/ds18b20"
	_ "github.com/zlowred/embd/host/rpi"
)

func main() {
	if err := embd.InitW1(); err != nil {
		panic(err)
	}
	defer embd.CloseW1()

	w1 := embd.NewW1Bus(0)

	devs, err := w1.ListDevices()

	if err != nil {
		panic(err)
	}

	var name string = ""
	for _, dev := range devs {
		if strings.HasPrefix(dev, "28-") {
			name = dev
			break
		}
	}

	if name == "" {
		fmt.Println("No DS18B20 devices found")
	}

	fmt.Printf("Using DS18B20 device %s\n", name)
	w1d, err := w1.Open(name)

	if err != nil {
		panic(err)
	}

	sensor := ds18b20.New(w1d)

	err = sensor.SetResolution(ds18b20.Resolution_12bit)

	if err != nil {
		panic(err)
	}

	err = sensor.ReadTemperature()

	if err != nil {
		panic(err)
	}

	fmt.Printf("Measured temperature: %vC\n", sensor.Celsius())
	fmt.Printf("Measured temperature: %vF\n", sensor.Fahrenheit())

	err = sensor.SetResolution(ds18b20.Resolution_9bit)

	if err != nil {
		panic(err)
	}

	err = sensor.ReadTemperature()

	if err != nil {
		panic(err)
	}

	fmt.Printf("Measured temperature: %vC\n", sensor.Celsius())
	fmt.Printf("Measured temperature: %vF\n", sensor.Fahrenheit())
}
