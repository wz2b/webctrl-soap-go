package alcsoap

import "encoding/xml"

type AlcFaultEnvelope struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Body    AlcFaultBody
}

type AlcFaultBody struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`
	Fault   AlcError `xml:"Fault"`
}

// AlcError is called "Fault" in the SOAP API but it is named Error here so it is recognizable as a go error type.
type AlcError struct {
	FaultCode   string         `xml:"http://xml.apache.org/axis/ faultcode"`
	FaultString string         `xml:"faultstring"`
	Detail      AlcFaultDetail `xml:"detail"`
}

type AlcFaultDetail struct {
	Hostname string `xml:"http://xml.apache.org/axis/ hostname"`
}

func (e AlcError) Error() string {
	return e.FaultCode + ": " + e.FaultString
}
