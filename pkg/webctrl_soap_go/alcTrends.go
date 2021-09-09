package webctrl_soap_go

import (
	"encoding/xml"
	"strconv"
	"time"
)

var loc, _ = time.LoadLocation("America/New_York")

type alcGetTrendDataRequest struct {
	XMLName        xml.Name `xml:"alcsoap:getTrendData"`
	Path           string   `xml:"alcsoap:trendLogPath"`
	Start          string   `xml:"alcsoap:sTime"`
	End            string   `xml:"alcsoap:eTime"`
	LimitFromStart bool     `xml:"alcsoap:limitFromStart"`
	MaxRecords     int      `xml:"alcsoap:maxRecords"`
}

type alcGetTrendDataResponseEnvelope struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Body    *alcGetTrendDataResponseBody
}

type alcGetTrendDataResponseBody struct {
	XMLName  xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`
	Response *alcGetTrendDataResponse
}

type alcGetTrendDataResponse struct {
	XMLName xml.Name `xml:"http://soap.core.green.controlj.com getTrendDataResponse"`
	Return  *alcGetTrendDataReturn
}

type alcGetTrendDataReturn struct {
	XMLName xml.Name `xml:"getTrendDataReturn"`
	Points  []string `xml:"getTrendDataReturn"`
}

type TrendPoint struct {
	Time        time.Time
	TimeString  string
	ValueString string
	Value       float64
}

func (this *TrendPoint) IsValid() bool {
	return this.ValueString != ""
}

func (this *TrendService) getTrendDataChunk(gql string, startTime time.Time, endTime time.Time, fromStart bool, maxRecords int) ([]TrendPoint, error) {
	request := alcEnvelope{
		XMLNsSoap: "http://schemas.xmlsoap.org/soap/envelope/",
		XMLNsAlc:  "http://soap.core.green.controlj.com",
		Body: alcBody{
			Payload: alcGetTrendDataRequest{
				Path:           gql,
				Start:          TimeToAlcFormat(startTime.Local()),
				End:            TimeToAlcFormat(endTime.Local()),
				LimitFromStart: fromStart,
				MaxRecords:     maxRecords,
			},
		},
	}

	requestPayload, err := xml.Marshal(request)
	if err != nil {
		return nil, err
	}

	response, err := call(this.Endpoint, this.User, this.password, xmlheader+string(requestPayload))

	if err != nil {
		//log.Println("Failure getting response ", err)
		return nil, err
	}

	respObj := new(alcGetTrendDataResponseEnvelope)
	err = xml.Unmarshal(response, respObj)
	if err != nil {
		return nil, err
	}

	rawPoints := respObj.Body.Response.Return.Points

	numRawPoints := len(rawPoints) / 2

	points := make([]TrendPoint, 0, numRawPoints)

	for i := 0; i < numRawPoints; i++ {
		timeString := rawPoints[i*2]
		valueString := rawPoints[(i*2)+1]

		tm, err := ParseAlcTime(timeString)

		if err != nil {
			return nil, err
		}

		val, err := strconv.ParseFloat(valueString, 32)
		if err != nil {
			//
			// Occasionally a value will come through that can't be parsed.  These are
			// generally various "other items" in the trend log, including things like
			// time synchronization messages.  The ALC internal interface lets you distinguish
			// these, but the SOAP interface does not.  If something can't be parsed
			// as a float, simply ignore it.
			//
			//backend.Logger.Debug("Unable to parse value ", valueString, ": ", err)
			continue
		}

		point := TrendPoint{
			TimeString:  timeString,
			ValueString: valueString,
			Time:        tm,
			Value:       float64(val),
		}

		points = append(points, point)
	}

	return points, nil
}
