package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/alexcesaro/statsd"
	"github.com/goburrow/modbus"
)

func main() {
	server := flag.String("statsd-server", "", "statsd UDP server")
	flag.Parse()

	c, err := statsd.New(
		statsd.Address(*server),
		statsd.Prefix("solar"),
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer c.Close()

	handler := modbus.NewRTUClientHandler("/dev/ttyUSB0")
	handler.BaudRate = 9600
	handler.DataBits = 8
	handler.Parity = "N"
	handler.StopBits = 2
	handler.SlaveId = 1
	handler.Timeout = 5 * time.Second

	if err := handler.Connect(); err != nil {
		fmt.Println("error conecting: ", err)
		os.Exit(1)
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

	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt)

	for {
		select {
		case <-s:
			fmt.Println("shutting down")
			return
		default:
			// modbus uses 16bit registers. Results is a slice of uint8. Every two
			// registers need to be combined into a uint16 value for their real
			// value.
			//
			// Get value for Adc_vb_f
			hb := results[0]
			lb := results[1]
			b := uint16(uint16(hb)<<8 | uint16(lb))
			v := float64(b) * conv
			c.Gauge("adcvbf", v)
			//fmt.Printf("Adc_vb_f=%f\n", v)

			// Get value for Adc_va_f
			hb = results[2]
			lb = results[3]
			b = uint16(uint16(hb)<<8 | uint16(lb))
			v = float64(b) * conv
			c.Gauge("adcvaf", v)
			//fmt.Printf("Adc_va_f=%f\n", v)

			// Get value for Ahc_daily
			hb = results[74]
			lb = results[75]
			b = uint16(uint16(hb)<<8 | uint16(lb))
			v = float64(b) * 0.1
			c.Gauge("ahcdaily", v)
			//fmt.Printf("Ahc_daily=%f\n", v)

			// Get value for Ahl_daily
			hb = results[76]
			lb = results[77]
			b = uint16(uint16(hb)<<8 | uint16(lb))
			v = float64(b) * 0.1
			c.Gauge("ahldaily", v)
			//fmt.Printf("Ahl_daily=%f\n", v)
		}

		time.Sleep(5 * time.Second)
	}
}
