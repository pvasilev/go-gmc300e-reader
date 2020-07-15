package main

import (
	"GMC300EReader/gmc300e"
	"go.bug.st/serial.v1"
	"log"
	"time"
)

func main() {
	config := new(gmc300e.ConnectorConfig)
	config.UsbPort = "/dev/ttyUSB0"
	config.BaudRate = 57600
	config.DataBits = 8
	config.Parity = serial.NoParity
	config.StopBits = serial.OneStopBit
	config.ReadTimeout = 3
	config.WriteTimeout = 3
	config.WaitBetweenSendAndReceive = 500 * time.Millisecond

	connector := new(gmc300e.Connector)
	err := connector.Connect(*config)
	if err != nil {
		log.Fatalf("failed to open port %s due to %s", config.UsbPort, err)
	}
	defer connector.Disconnect()

	connector.EnumeratePorts()
	ver, err := connector.GetVer()
	if err != nil {
		log.Fatalf("failed to get version due to %s\n", err)
	}
	log.Printf("Got version: %s\n", ver)

	cpm, err := connector.GetCpm()
	if err != nil {
		log.Fatalf("failed to get CPM due to %s\n", err)
	}
	log.Printf("Got CPM: %d\n", cpm)

	cpml, err := connector.GetCpml()
	if err != nil {
		log.Fatalf("failed to get CPML due to %s\n", err)
	}
	log.Printf("Got CPML: %d\n", cpml)

	cpmh, err := connector.GetCpmh()
	if err != nil {
		log.Fatalf("failed to get CPMH due to %s\n", err)
	}
	log.Printf("Got CPMH: %d\n", cpmh)

	cps, err := connector.GetCps()
	if err != nil {
		log.Fatalf("failed to get CPS due to %s\n", err)
	}
	log.Printf("Got CPS: %d\n", cps)

	cpsl, err := connector.GetCpsl()
	if err != nil {
		log.Fatalf("failed to get CPSL due to %s\n", err)
	}
	log.Printf("Got CPML: %d\n", cpsl)

	cpsh, err := connector.GetCpsh()
	if err != nil {
		log.Fatalf("failed to get CPSH due to %s\n", err)
	}
	log.Printf("Got CPMH: %d\n", cpsh)

	cfg, err := connector.GetCfg()
	if err != nil {
		log.Fatalf("failed to get config due to %s\n", err)
	}
	log.Printf("Got CFG: %q\n", cfg)

	dateTime, err := connector.GetDateTime()
	if err != nil {
		log.Fatalf("failed to get datetime due to %s\n", err)
	}
	log.Printf("Got DateTime: %q\n", dateTime)


}
