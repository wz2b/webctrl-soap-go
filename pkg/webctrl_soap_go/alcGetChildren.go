package webctrl_soap_go

import (
	"encoding/xml"
	"log"
)

type alcGetChildrenRequest struct {
	XMLName    xml.Name `xml:"alcsoap:getFilteredChildren"`
	Expression string   `xml:"alcsoap:expression"`
	Filter     string   `xml:"alcsoap:filter"`
}

type alcGetChildrenResponseEnvelope struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Body    alcGetChildrenResponseBody
}

type alcGetChildrenResponseBody struct {
	XMLName xml.Name  `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`
	Nodes   []GqlNode `xml:"multiRef"`
}

type GqlNode struct {
	XMLName       xml.Name `xml:"multiRef"`
	ReferenceName string   `xml:"referenceName"`
	DisplayName   string   `xml:"displayName"`
	Type          string   `xml:"type"`
}

func (this *EvalService) GetChildren(gql string, filter func(GqlNode) bool) ([]GqlNode, error) {
	request := alcEnvelope{
		XMLNsSoap: "http://schemas.xmlsoap.org/soap/envelope/",
		XMLNsAlc:  "http://soap.core.green.controlj.com",
		Body: alcBody{
			Payload: alcGetChildrenRequest{
				Expression: gql,
				Filter:     "WEB_GEO",
			},
		},
	}
	payload, err := xml.Marshal(request)

	if err != nil {
		return nil, err
	}

	response, err := call(this.parent.httpClient, this.Endpoint, this.User, this.password, xmlheader+string(payload))

	if err != nil {
		log.Fatal("Failure", err)
	}
	//fmt.Println(xmlfmt.FormatXML(string(response), "", "  "))
	respObj := new(alcGetChildrenResponseEnvelope)

	err = xml.Unmarshal(response, respObj)

	if err != nil {
		return nil, err
	}

	nodes := respObj.Body.Nodes

	if filter != nil {
		n := 0
		for _, x := range nodes {
			if filter(x) {
				nodes[n] = x
				n++
			}
		}
		nodes = nodes[:n]
	}

	return nodes, nil
}
