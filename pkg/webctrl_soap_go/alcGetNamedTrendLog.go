package webctrl_soap_go

import "encoding/xml"

type alcGetNamedTrendLogRequest struct {
	XMLName          xml.Name `xml:"alcsoap:getNamedTrendLog"`
	EquipmentRefPath string   `xml:"alcsoap:eqRefPath"`
	TrendLogRefName  string   `xml:"alcsoap:trendLogRefName"`
}

type alcGetNamedTrendLogResponseEnvelope struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Body    alcGetNamedTrendLogResponseBody
}

type alcGetNamedTrendLogResponseBody struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`
	Node    GqlNode  `xml:"namedTrendLogResponse"`
}

func (this *EvalService) GetNamedTrendLog(equipment string, trend string) (*GqlNode, error) {
	request := alcEnvelope{
		XMLNsSoap: "http://schemas.xmlsoap.org/soap/envelope/",
		XMLNsAlc:  "http://soap.core.green.controlj.com",
		Body: alcBody{
			Payload: alcGetNamedTrendLogRequest{
				EquipmentRefPath: equipment,
				TrendLogRefName:  trend,
			},
		},
	}

	payload, err := xml.Marshal(request)
	if err != nil {
		return nil, err
	}

	response, err := call(this.Endpoint, this.User, this.password, xmlheader+string(payload))
	respObj := new(alcGetNamedTrendLogResponseEnvelope)
	err = xml.Unmarshal(response, respObj)
	return &respObj.Body.Node, err
}
