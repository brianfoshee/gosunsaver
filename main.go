package main

import (
	"fmt"
	"time"

	"github.com/goburrow/modbus"
)

func main() {
	handler := modbus.NewRTUClientHandler("/dev/ttyUSB0")
	handler.BaudRate = 9600
	handler.DataBits = 8
	handler.Parity = "N"
	handler.StopBits = 2
	handler.SlaveId = 1
	handler.Timeout = 5 * time.Second

	if err := handler.Connect(); err != nil {
		fmt.Println("error conecting: ", err)
		return
	}

	defer handler.Close()

	client := modbus.NewClient(handler)

	results, err := client.ReadHoldingRegisters(8, 44)
	if err != nil {
		fmt.Println("error reading registers:", err)
		return
	}

	fb := 32768.0
	conv := 100.0 / fb

	// modbus uses 16bit registers. Results is a slice of uint8. Every two
	// registers need to be combined into a uint16 number for their real
	// value.
	hb := results[0]
	lb := results[1]
	b := uint16(uint16(hb)<<8 | uint16(lb))

	fmt.Printf("Adc_vb_f=%f\n", float64(b)*conv)
}
