package internet

import (
	"errors"
	"net/url"

	"github.com/DeNetPRO/turbo-upnp/soap"
)

const (
	URN_LANDevice_1           = "urn:schemas-upnp-org:device:LANDevice:1"
	URN_WANConnectionDevice_1 = "urn:schemas-upnp-org:device:WANConnectionDevice:1"
	URN_WANDevice_1           = "urn:schemas-upnp-org:device:WANDevice:1"

	URN_WANIPConnection_1 = "urn:schemas-upnp-org:service:WANIPConnection:1"

	AddPortMappingAction       = "AddPortMapping"
	DeletePortMappingAction    = "DeletePortMapping"
	GetExternalIPAddressAction = "GetExternalIPAddress"
)

func AddPortMapping(newRemoteHost, newInternalClient, newProtocol, newPortMappingDescription string, newExternalPort, newInternalPort uint16, newEnabled bool, newLeaseDuration uint32, clientURL *url.URL) error {
	if clientURL == nil {
		return errors.New("url is empty")
	}

	exPort, err := soap.MarshalU16(newExternalPort)
	if err != nil {
		return err
	}

	inPort, err := soap.MarshalU16(newInternalPort)
	if err != nil {
		return err
	}

	duration, err := soap.MarshalU32(newLeaseDuration)
	if err != nil {
		return err
	}

	params := map[string]string{
		"NewRemoteHost":             newRemoteHost,
		"NewExternalPort":           exPort,
		"NewProtocol":               newProtocol,
		"NewInternalPort":           inPort,
		"NewInternalClient":         newInternalClient,
		"NewEnabled":                soap.MarshalBoolean(newEnabled),
		"NewPortMappingDescription": newPortMappingDescription,
		"NewLeaseDuration":          duration,
	}

	action := &soap.SoapAction{
		Params: params,
	}

	return soap.PerformAction(AddPortMappingAction, clientURL, action, nil)
}

func DeletePortMapping(newRemoteHost string, newExternalPort uint16, newProtocol string, clientURL *url.URL) error {
	if clientURL == nil {
		return errors.New("url is empty")
	}

	port, err := soap.MarshalU16(newExternalPort)
	if err != nil {
		return err
	}

	params := map[string]string{
		"NewRemoteHost":   newRemoteHost,
		"NewExternalPort": port,
		"NewProtocol":     newProtocol,
	}

	action := &soap.SoapAction{
		Params: params,
	}

	return soap.PerformAction(DeletePortMappingAction, clientURL, action, nil)
}

func GetExternalIPAddress(clientURL *url.URL) (string, error) {
	if clientURL == nil {
		return "", errors.New("url is empty")
	}

	type responseAction struct {
		NewExternalIPAddress string
	}

	response := &responseAction{}

	if err := soap.PerformAction(GetExternalIPAddressAction, clientURL, nil, response); err != nil {
		return "", err
	}

	return response.NewExternalIPAddress, nil
}
