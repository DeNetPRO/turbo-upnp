package upnp

import (
	"errors"
	"log"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/DeNetPRO/turbo-upnp/internet"
	"github.com/DeNetPRO/turbo-upnp/ssdp"
)

const (
	clientURLEnd = "/ctl/IPConn"
)

//Device contains real location your router (ip address)
//and clientURL. ClientURl is address, where we will send requests
type Device struct {
	location  string
	clientURL *url.URL
}

//Initialisation device, which is detected by ssdp
func InitDevice() (*Device, error) {
	device := &Device{}
	times := 3
	timeSleep := time.Millisecond * 500
	for i := 0; i < times; i++ {
		list, err := ssdp.Search(ssdp.All, 1, "")
		if err != nil {
			return nil, err
		}

		var address string
		for _, srv := range list {
			if srv.Type == internet.URN_WANIPConnection_1 {
				address = srv.Location
			}
		}

		if address == "" {
			time.Sleep(timeSleep)
			continue
		}

		addressSplit := strings.Split(address, "http://")

		if len(addressSplit) != 2 {
			return nil, errors.New("invalid address")
		}

		ipSplit := strings.Split(addressSplit[1], "/")

		if len(ipSplit) != 2 {
			return nil, errors.New("invalid address")
		}

		device.location = ipSplit[0]
		clientAddress := "http://" + device.location + clientURLEnd
		clientUrl, err := url.Parse(clientAddress)
		if err != nil {
			return device, err
		}

		device.clientURL = clientUrl

		return device, nil
	}

	return nil, errors.New("no available internet gateway devices")
}

//Internet Gateway Device address
func (d *Device) Location() string {
	return d.location
}

//Real public IP address
func (d *Device) PublicIP() (string, error) {
	return internet.GetExternalIPAddress(d.clientURL)
}

//Forward opens ports on the router
func (d *Device) Forward(port int, mapDescription string) error {
	return internet.AddPortMapping("", getInternalIP(), "TCP", mapDescription, uint16(port), uint16(port), true, 0, d.clientURL)
}

//Close ports on the router
func (d *Device) Close(port int) error {
	return internet.DeletePortMapping("", uint16(port), "TCP", d.clientURL)
}

func getInternalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}
