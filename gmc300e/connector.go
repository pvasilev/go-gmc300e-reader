package gmc300e

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"go.bug.st/serial.v1"
	"go.bug.st/serial.v1/enumerator"
	"log"
	"time"
)

type ConnectorConfig struct {
	UsbPort string
	BaudRate int
	DataBits int
	StopBits serial.StopBits
	Parity serial.Parity
	ReadTimeout int
	WriteTimeout int
	WaitBetweenSendAndReceive time.Duration
}

type Connector struct {
	config   ConnectorConfig
	isOpened bool
	port     serial.Port
}

func (c *Connector) openPort(portName string, mode *serial.Mode) (serial.Port, error) {
	return serial.Open(portName, mode)
}

func (c *Connector) closePort(port serial.Port) error {
	return port.Close()
}

func (c *Connector) Connect(config ConnectorConfig) error {
	portMode := serial.Mode{
		BaudRate: config.BaudRate,
		DataBits: config.DataBits,
		Parity:   config.Parity,
		StopBits: config.StopBits,
	}
	port, err := c.openPort(config.UsbPort, &portMode)
	if err != nil {
		c.isOpened = false
		return err
	}
	c.port = port
	c.isOpened = true
	c.config = config
	return nil
}

func (c *Connector) Disconnect() error {
	return c.closePort(c.port)
}

func (c *Connector) ReadFromPort() ([]byte, error) {
	if ! c.isOpened {
		return nil, errors.New("port is not opened")
	}
	var buffer bytes.Buffer
	tempBuff := make([]byte,1024)

	bits, err := c.port.GetModemStatusBits()
	if err != nil {
		log.Printf("Failed to get serial line status due to %s", err)
		return nil, err
	}
	log.Printf("Bits of serial line status: %q\n", bits)

	n, err := c.port.Read(tempBuff)
	log.Printf("%d bytes read from port\n", n)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		fmt.Println("End of stream detected")
	}
	buffer.Write(tempBuff[:n])

	return buffer.Bytes(), nil
}

func (c *Connector) WriteToPort(data []byte) error {
	startIdx := 0
	endIdx := len(data)
	for startIdx, err := c.port.Write(data[startIdx:endIdx]); startIdx < endIdx; {
		if err != nil {
			log.Printf("Failed to write to port %s due to %s", c.config.UsbPort, err)
			return err
		}
	}
	return nil
}

func (c *Connector) SendCommandAndGetResponse(command string) ([]byte, error) {
	log.Printf("Sending command %s...\n", command)
	err := c.WriteToPort([]byte(command))
	if err != nil {
		log.Printf("Failed to send %s to port %s\n", command, c.config.UsbPort)
		return nil, err
	}
	log.Printf("Reading the command %s response...\n", command)
	time.Sleep(c.config.WaitBetweenSendAndReceive)
	response, err := c.ReadFromPort()
	if err != nil {
		log.Printf("Failed to read response to the %s command from port %s\n", command, c.config.UsbPort)
		return nil, err
	}
	log.Printf("Got %d bytes as response to command: %q\n", len(response), response)

	return response, nil
}

func (c *Connector) SendCommandAndGetResponseAsString(command string) (string, error) {
	responseAsByteArray, err := c.SendCommandAndGetResponse(command)
	if err != nil {
		return "", err
	}
	if responseAsByteArray == nil || len(responseAsByteArray) == 0 {
		return "", nil
	}
	return string(responseAsByteArray), nil
}

func (c *Connector) SendCommandAndGetResponseAsDateTime(command string) (time.Time, error) {
	responseAsByteArray, err := c.SendCommandAndGetResponse(command)
	if err != nil {
		return time.Time{}, err
	}
	if responseAsByteArray == nil || len(responseAsByteArray) == 0 {
		return time.Time{}, nil
	}
	if len(responseAsByteArray) != 7 {
		return time.Time{}, errors.New(fmt.Sprintf("invalid Date/Time response array length. Expected 7, got %d", len(responseAsByteArray)))
	}
	year := 2000 + int(responseAsByteArray[0])
	month := time.Month(responseAsByteArray[1])
	day := int(responseAsByteArray[2])
	hour := int(responseAsByteArray[3])
	min := int(responseAsByteArray[4])
	sec := int(responseAsByteArray[5])
	result := time.Date(year, month, day, hour, min, sec, 0, time.Local)
	return result, nil
}

func (c *Connector) SendCommandAndGetResponseAsUint16(command string) (uint16, error) {
	responseAsByteArray, err := c.SendCommandAndGetResponse(command)
	if err != nil {
		return 0, err
	}
	if responseAsByteArray == nil || len(responseAsByteArray) == 0 {
		return 0, nil
	}
	return binary.BigEndian.Uint16(responseAsByteArray), nil
}

func (c *Connector) SendCommandAndGetResponseAsUint32(command string) (uint32, error) {
	responseAsByteArray, err := c.SendCommandAndGetResponse(command)
	if err != nil {
		return 0, err
	}
	if responseAsByteArray == nil || len(responseAsByteArray) == 0 {
		return 0, nil
	}
	return binary.BigEndian.Uint32(responseAsByteArray), nil
}

func (c *Connector) SendCommandAndGetResponseAsUint64(command string) (uint64, error) {
	responseAsByteArray, err := c.SendCommandAndGetResponse(command)
	if err != nil {
		return 0, err
	}
	if responseAsByteArray == nil || len(responseAsByteArray) == 0 {
		return 0, nil
	}
	return binary.BigEndian.Uint64(responseAsByteArray), nil
}

func (c *Connector) EnumeratePorts() error {
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return err
	}
	if len(ports) == 0 {
		log.Print("No ports could be enumerated")
		return nil
	}
	for _, port := range ports {
		fmt.Printf("Found port: %s\n", port.Name)
		if port.IsUSB {
			fmt.Printf("\tUSB ID:\t%s:%s\n", port.VID, port.PID)
			fmt.Printf("\tUSB Serial:\t%s\n", port.SerialNumber)
		}
	}
	return nil
}

func (c *Connector) constructCommand(command string) string {
	return "<"+command+">>"
}

func (c *Connector) GetVer() (string, error) {
	return c.SendCommandAndGetResponseAsString(c.constructCommand("GETVER"))
}

func (c *Connector) GetCpm() (uint16, error) {
	return c.SendCommandAndGetResponseAsUint16(c.constructCommand("GETCPM"))
}

func (c *Connector) GetCpml() (uint16, error) {
	return c.SendCommandAndGetResponseAsUint16(c.constructCommand("GETCPML"))
}

func (c *Connector) GetCpmh() (uint16, error) {
	return c.SendCommandAndGetResponseAsUint16(c.constructCommand("GETCPMH"))
}

func (c *Connector) GetCps() (uint16, error) {
	return c.SendCommandAndGetResponseAsUint16(c.constructCommand("GETCPS"))
}

func (c *Connector) GetCpsh() (uint16, error) {
	return c.SendCommandAndGetResponseAsUint16(c.constructCommand("GETCPSH"))
}

func (c *Connector) GetCpsl() (uint16, error) {
	return c.SendCommandAndGetResponseAsUint16(c.constructCommand("GETCPSL"))
}

func (c *Connector) GetCfg() (string, error) {
	return c.SendCommandAndGetResponseAsString(c.constructCommand("GETCFG"))
}

func (c *Connector) GetDateTime() (time.Time, error) {
	return c.SendCommandAndGetResponseAsDateTime(c.constructCommand("GETDATETIME"))
}


