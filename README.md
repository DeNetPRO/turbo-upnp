## Installation
Remember to enable UPnP on your router!

```bash
go get -u github.com/DeNetPRO/turbo-upnp
```

## Usage
```golang
import upnp "github.com/DeNetPRO/turbo-upnp"

//initializes internet gateway device
device := upnp.InitDevice()

// forward port
description := "test"
port := 9999
device.Forward(port, description)
```
![Image alt](./ipv4_igd.png)

```golang
// close port 
device.Close(port)

// get your public ip 
ip := device.PublicIP()
```