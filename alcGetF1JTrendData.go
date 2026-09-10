package alcsoap

import "encoding/xml"

type alcGetF1JrendDataRequest struct {
	XMLName        xml.Name `xml:"alcsoap:getF1JTrendData"`
	Path           string   `xml:"alcsoap:trendLogPath"`
	Start          int32    `xml:"alcsoap:startTime"`
	End            int32    `xml:"alcsoap:endTime"`
	LimitFromStart bool     `xml:"alcsoap:limitFromStart"`
	MaxRecords     int      `xml:"alcsoap:maxRecords"`
}

type alcGetF1JTrendDataResponseEnvelope struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Body    *alcGetF1JTrendDataResponseBody
}

type alcGetF1JTrendDataResponseBody struct {
	XMLName  xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`
	Response *alcGetF1JTrendDataResponse
}

type alcGetF1JTrendDataResponse struct {
	XMLName xml.Name `xml:"http://soap.core.green.controlj.com getF1JTrendDataResponse"`
	Return  *alcGetF1JTrendDataReturn
}

type alcGetF1JTrendDataReturn struct {
	XMLName xml.Name          `xml:"getF1JTrendDataReturn"`
	Points  []CoreTrendRecord `xml:"getF1JTrendDataReturn"`
}

type CoreTrendRecord struct {
}
