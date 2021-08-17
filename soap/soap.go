package soap

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
)

const (
	soapEncodingStyle = "http://schemas.xmlsoap.org/soap/encoding/"
	soapPrefix        = xml.Header + `<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/" s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/"><s:Body>`
	soapSuffix        = `</s:Body></s:Envelope>`
)

const URN_WANPPPConnection_1 = "urn:schemas-upnp-org:service:WANPPPConnection:1"

var xmlCharRx = regexp.MustCompile("[<>&]")

type SoapAction struct {
	Params map[string]string
}

type soapEnvelope struct {
	XMLName       xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	EncodingStyle string   `xml:"http://schemas.xmlsoap.org/soap/envelope/ encodingStyle,attr"`
	Body          soapBody `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`
}

type soapBody struct {
	Fault     *SOAPFaultError `xml:"Fault"`
	RawAction []byte          `xml:",innerxml"`
}

// SOAPFaultError implements error, and contains SOAP fault information.
type SOAPFaultError struct {
	FaultCode   string `xml:"faultCode"`
	FaultString string `xml:"faultString"`
	Detail      struct {
		Raw []byte `xml:",innerxml"`
	} `xml:"detail"`
}

// PerformSOAPAction makes a SOAP request, with the given action.
func PerformAction(actionName string, url *url.URL, action *SoapAction, responseAction interface{}) error {
	requestBytes, err := encodeRequestAction(actionName, action)
	if err != nil {
		return err
	}

	client := &http.Client{}
	request := &http.Request{
		Method: "POST",
		URL:    url,
		Header: http.Header{
			"SOAPACTION":   []string{`"` + URN_WANPPPConnection_1 + "#" + actionName + `"`},
			"CONTENT-TYPE": []string{"text/xml; charset=\"utf-8\""},
		},
		Body:          io.NopCloser(bytes.NewBuffer(requestBytes)),
		ContentLength: int64(len(requestBytes)),
	}

	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("error performing SOAP HTTP request: %v", err)
	}

	if response != nil {
		defer response.Body.Close()
	}

	if response.StatusCode != 200 && response.ContentLength == 0 {
		return fmt.Errorf("SOAP request got HTTP %s", response.Status)
	}

	responseEnv := newSOAPEnvelope()
	decoder := xml.NewDecoder(response.Body)
	if err := decoder.Decode(responseEnv); err != nil {
		return fmt.Errorf("error decoding response body: %v", err)
	}

	if responseEnv.Body.Fault != nil {
		return responseEnv.Body.Fault
	} else if response.StatusCode != 200 {
		return fmt.Errorf("SOAP request got HTTP %s", response.Status)
	}

	if responseAction != nil {
		if err := xml.Unmarshal(responseEnv.Body.RawAction, responseAction); err != nil {
			return fmt.Errorf("error unmarshalling out action: %v, %v", err, responseEnv.Body.RawAction)
		}
	}

	return nil
}

// newSOAPAction creates a soapEnvelope with the given action and arguments.
func newSOAPEnvelope() *soapEnvelope {
	return &soapEnvelope{
		EncodingStyle: soapEncodingStyle,
	}
}

// encodeRequestAction is a hacky way to create an encoded SOAP envelope
// containing the given action. Experiments with one router have shown that it
// 500s for requests where the outer default xmlns is set to the SOAP
// namespace, and then reassigning the default namespace within that to the
// service namespace. Hand-coding the outer XML to work-around this.
func encodeRequestAction(actionName string, request *SoapAction) ([]byte, error) {
	requestBuf := new(bytes.Buffer)
	requestBuf.WriteString(soapPrefix)
	requestBuf.WriteString(`<u:`)
	xml.EscapeText(requestBuf, []byte(actionName))
	requestBuf.WriteString(` xmlns:u="`)
	xml.EscapeText(requestBuf, []byte(URN_WANPPPConnection_1))
	requestBuf.WriteString(`">`)
	if request != nil {
		if err := encodeRequestArgs(requestBuf, request); err != nil {
			return nil, err
		}
	}
	requestBuf.WriteString(`</u:`)
	xml.EscapeText(requestBuf, []byte(actionName))
	requestBuf.WriteString(`>`)
	requestBuf.WriteString(soapSuffix)
	return requestBuf.Bytes(), nil
}

func encodeRequestArgs(w *bytes.Buffer, action *SoapAction) error {
	enc := xml.NewEncoder(w)

	for field, value := range action.Params {
		elem := xml.StartElement{xml.Name{"", field}, nil}
		if err := enc.EncodeToken(elem); err != nil {
			return fmt.Errorf("error encoding start element for SOAP arg %q: %v", field, err)
		}

		if err := enc.Flush(); err != nil {
			return fmt.Errorf("error flushing start element for SOAP arg %q: %v", field, err)
		}

		if _, err := w.Write([]byte(escapeXMLText(value))); err != nil {
			return fmt.Errorf("error writing value for SOAP arg %q: %v", field, err)
		}

		if err := enc.EncodeToken(elem.End()); err != nil {
			return fmt.Errorf("error encoding end element for SOAP arg %q: %v", field, err)
		}
	}

	return enc.Flush()
}

func escapeXMLText(s string) string {
	return xmlCharRx.ReplaceAllStringFunc(s, replaceEntity)
}

func replaceEntity(s string) string {
	switch s {
	case "<":
		return "&lt;"
	case ">":
		return "&gt;"
	case "&":
		return "&amp;"
	}
	return s
}

func (err *SOAPFaultError) Error() string {
	return fmt.Sprintf("SOAP fault: %s", err.FaultString)
}

func MarshalBoolean(v bool) string {
	if v {
		return "1"
	}
	return "0"
}

func MarshalU16(v uint16) (string, error) {
	return strconv.FormatUint(uint64(v), 10), nil
}

func MarshalU32(v uint32) (string, error) {
	return strconv.FormatUint(uint64(v), 10), nil
}
