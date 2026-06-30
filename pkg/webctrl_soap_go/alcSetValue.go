package webctrl_soap_go

import (
	"encoding/xml"
)

type alcSetValueRequest struct {
	XMLName      xml.Name `xml:"setValue"`
	Expression   string   `xml:"alcsoap:expression"`
	NewValue     string   `xml:"alcsoap:newValue"`
	ChangeReason string   `xml:"alcsoap:changeReason"`
}

type alcSetValueResponseEnvelope struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Body    alcSetValueResponseBody
}

type alcSetValueResponseBody struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`
}

func (this *EvalService) SetValue(gql string, newValue string, changeReason string) error {
	request := alcEnvelope{
		XMLNsSoap: "http://schemas.xmlsoap.org/soap/envelope/",
		XMLNsAlc:  "http://soap.core.green/controlj.com",
		Body: alcBody{
			Payload: alcSetValueRequest{
				Expression:   gql,
				NewValue:     newValue,
				ChangeReason: changeReason,
			},
		},
	}
	payload, err := xml.Marshal(request)

	if err != nil {
		return err
	}

	_, err = call(this.Endpoint, this.User, this.password, xmlheader+string(payload))
	if err != nil {
		return err
	}

	return nil
}
