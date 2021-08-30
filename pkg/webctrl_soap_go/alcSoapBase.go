package webctrl_soap_go

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const xmlheader = "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n"

var debugExec = false

type alcEnvelope struct {
	XMLName   xml.Name `xml:"soap:Envelope"`
	XMLNsSoap string   `xml:"xmlns:soap,attr"`
	XMLNsAlc string   `xml:"xmlns:alcsoap,attr"`
	Body     alcBody
}

type alcBody struct {
	XMLName xml.Name `xml:"soap:Body"`
	Payload interface{}
}

var client = http.Client{
	Timeout: 30 * time.Second,
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

type SoapService struct {
	Eval  EvalService
	Trend TrendService
}

type EvalService struct {
	User     string
	password string
	Endpoint string
}

type TrendService struct {
	User     string
	password string
	Endpoint string
}

func NewSoapService(host string, user string, password string) *SoapService {
	if !strings.HasSuffix(host, "/") {
		host = host + "/"
	}

	return &SoapService{
		Eval:  EvalService{Endpoint: host + "_common/webservices/Eval", User: user, password: password},
		Trend: TrendService{Endpoint: host + "_common/webservices/Trend", User: user, password: password},
	}
}

func call(endpoint string, username string, password string, payload string) ([]byte, error) {
	var err error
	request, err := http.NewRequest(http.MethodPost,
		endpoint,
		bytes.NewBufferString(payload))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Accept", "text/xml, multipart/related")
	request.Header.Set("Content-Type", "text/xml; charset=utf-8")
	request.Header.Set("SOAPAction", "grafana-alcsoap")
	request.Header.Set("Authorization", basicAuth(username, password))

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, errors.New(response.Status)
	}

	if err != nil {
		return nil, err
	} else {
		bodyBytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}

		return bodyBytes, nil
	}
}
