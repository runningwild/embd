// +build ignore

package main

import (
	"flag"
	"fmt"

	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/all"
)

func main() {
	flag.Parse()

	if err := embd.InitI2C(); err != nil {
		panic(err)
	}
	defer embd.CloseI2C()

	bus := embd.NewI2CBus(1)

	sensor := npa700.New(bus)

	err := sensor.Read()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Temp is %fC\n", sensor.Celsius())
	fmt.Printf("Temp is %fF\n", sensor.Fahrenheit())
	fmt.Printf("Pres is %fPa\n", sensor.Pascals(0, 1638, 14745, -6894.76, 6894.76))
}
