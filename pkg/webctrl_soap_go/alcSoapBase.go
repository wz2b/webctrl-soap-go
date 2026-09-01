package webctrl_soap_go

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const xmlheader = "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n"

var debugExec = false

type alcEnvelope struct {
	XMLName   xml.Name `xml:"soap:Envelope"`
	XMLNsSoap string   `xml:"xmlns:soap,attr"`
	XMLNsAlc  string   `xml:"xmlns:alcsoap,attr"`
	Body      alcBody
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
	Name  string
	Eval  EvalService
	Trend TrendService

	httpClient *http.Client
}

type EvalService struct {
	User     string
	password string
	Endpoint string

	parent *SoapService
}

type TrendService struct {
	User     string
	password string
	Endpoint string

	parent    *SoapService
	chunkSize int
}

// NewSoapService creates a WebCTRL SOAP service using the default HTTP client.
//
// This is the normal constructor for callers that do not need custom HTTP or TLS
// behavior. The default client uses Go's standard HTTP transport and therefore
// uses the operating system's normal trusted certificate store.
//
// Use this when WebCTRL is using a certificate chain that the runtime already
// trusts.
//
// For environments with incomplete enterprise certificate chains, private CAs,
// custom timeouts, proxies, or other HTTP transport requirements, use
// NewSoapServiceWithHTTPClient instead.
func NewSoapService(host string, user string, password string) *SoapService {
	return NewSoapServiceWithHTTPClient(host, user, password, nil)
}

// NewSoapServiceWithHTTPClient creates a WebCTRL SOAP service using the provided
// HTTP client.
//
// This constructor exists so callers can control HTTP/TLS behavior without the
// library relying on package-global mutable state. In particular, it allows
// callers to provide an HTTP client that trusts additional CA certificates,
// uses a custom transport, sets different timeouts, uses a proxy, or applies
// other client-specific behavior.
//
// This was added because some WebCTRL installations do not present a complete
// publicly trusted certificate chain to all runtimes. For example, a local
// workstation may trust an intermediate certificate installed into the OS,
// while an AWS Lambda runtime may not. Passing a caller-owned HTTP client lets
// the application add the required CA certificate explicitly without weakening
// TLS validation globally or using InsecureSkipVerify.
//
// If httpClient is nil, a default HTTP client with a 30 second timeout is used.
//
// The supplied host may be passed with or without a trailing slash. The
// constructor normalizes it and builds the WebCTRL Eval and Trend SOAP endpoint
// URLs from it.
func NewSoapServiceWithHTTPClient(host string, user string, password string, httpClient *http.Client) *SoapService {
	if !strings.HasSuffix(host, "/") {
		host = host + "/"
	}

	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	service := &SoapService{
		Name:       host,
		httpClient: httpClient,
		Eval: EvalService{
			Endpoint: host + "_common/webservices/Eval",
			User:     user,
			password: password,
		},
		Trend: TrendService{
			Endpoint:  host + "_common/webservices/Trend",
			User:      user,
			password:  password,
			chunkSize: 2000,
		},
	}

	service.Eval.parent = service
	service.Trend.parent = service

	return service
}

func NewHTTPClientWithExtraCAPEM(extraCAs ...[]byte) (*http.Client, error) {
	rootCAs, err := x509.SystemCertPool()
	if rootCAs == nil || err != nil {
		rootCAs = x509.NewCertPool()
	}

	for i, pem := range extraCAs {
		if len(pem) == 0 {
			continue
		}
		if ok := rootCAs.AppendCertsFromPEM(pem); !ok {
			return nil, fmt.Errorf("failed to append extra CA certificate at index %d", i)
		}
	}

	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: rootCAs,
			},
		},
	}, nil
}

// func old_call(endpoint string, username string, password string, payload string) ([]byte, error) {
// 	var err error
// 	request, err := http.NewRequest(http.MethodPost,
// 		endpoint,
// 		bytes.NewBufferString(payload))
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	request.Header.Set("Accept", "text/xml, multipart/related")
// 	request.Header.Set("Content-Type", "text/xml; charset=utf-8")
// 	request.Header.Set("SOAPAction", "grafana-alcsoap")
// 	request.Header.Set("Authorization", basicAuth(username, password))
//
// 	response, err := client.Do(request)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	defer response.Body.Close()
//
// 	if response.StatusCode != 200 {
// 		bodyBytes, _ := ioutil.ReadAll(response.Body)
//
// 		faultObj := new(AlcFaultEnvelope)
// 		err = xml.Unmarshal(bodyBytes, faultObj)
// 		if err == nil {
// 			fault := faultObj.Body.Fault
// 			return nil, fault
// 		}
// 		return nil, errors.New(response.Status)
// 	}
//
// 	if err != nil {
// 		return nil, err
// 	} else {
// 		bodyBytes, err := ioutil.ReadAll(response.Body)
// 		if err != nil {
// 			return nil, err
// 		}
//
// 		return bodyBytes, nil
// 	}
// }

func call(httpClient *http.Client, endpoint string, username string, password string, payload string) ([]byte, error) {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	request, err := http.NewRequest(
		http.MethodPost,
		endpoint,
		bytes.NewBufferString(payload),
	)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Accept", "text/xml, multipart/related")
	request.Header.Set("Content-Type", "text/xml; charset=utf-8")
	request.Header.Set("SOAPAction", "grafana-alcsoap")
	request.Header.Set("Authorization", basicAuth(username, password))

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		faultObj := new(AlcFaultEnvelope)
		if err := xml.Unmarshal(bodyBytes, faultObj); err == nil {
			return nil, faultObj.Body.Fault
		}

		return nil, errors.New(response.Status)
	}

	return bodyBytes, nil
}
