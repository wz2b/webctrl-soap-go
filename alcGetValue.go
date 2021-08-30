package webctrl_soap_go

import (
	"encoding/xml"
)

type alcGetValueRequest struct {
	XMLName    xml.Name `xml:"getValue"`
	Expression string   `xml:"alcsoap:expression"`
}

type alcGetValueResponseEnvelope struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Body    alcGetValueResponseBody
}

type alcGetValueResponseBody struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`
	Value   string   `xml:"getValueResponse>getValueReturn"`
}

func (this *EvalService) GetValue(gql string) (string, error) {
	request := alcEnvelope{
		XMLNsSoap: "http://schemas.xmlsoap.org/soap/envelope/",
		XMLNsAlc:  "http://soap.core.green/controlj.com",
		Body: alcBody{
			Payload: alcGetValueRequest{
				Expression: gql,
			},
		},
	}
	payload, err := xml.Marshal(request)

	if err != nil {
		return "", err
	}

	response, err := call(this.Endpoint, this.User, this.password, xmlheader+string(payload))
	if err != nil {
		return "", err
	}

	respObj := new(alcGetValueResponseEnvelope)

	err = xml.Unmarshal(response, respObj)

	if err != nil {
		return "", err
	}

	return respObj.Body.Value, nil
}
