package alcsoap

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type alcGetF1JTrendDataRequest struct {
	XMLName        xml.Name `xml:"alcsoap:getF1JTrendData"`
	User           string   `xml:"alcsoap:user"`
	Passwd         string   `xml:"alcsoap:passwd"`
	Path           string   `xml:"alcsoap:trendLogPath"`
	Start          int64    `xml:"alcsoap:startTime"`
	End            int64    `xml:"alcsoap:endTime"`
	LimitFromStart bool     `xml:"alcsoap:limitFromStart"`
	MaxRecords     int      `xml:"alcsoap:maxRecords"`
}

//
// These names are deliberately different from the old/incomplete
// alcGetF1JTrendData.go types already in the package.
//

type f1jResponseEnvelope struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Body    *f1jResponseBody
}

type f1jResponseBody struct {
	XMLName  xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`
	Response *f1jResponse
	MultiRef []alcF1JMultiRef `xml:"multiRef"`
}

type f1jResponse struct {
	XMLName xml.Name   `xml:"getF1JTrendDataResponse"`
	Return  *f1jReturn `xml:"getF1JTrendDataReturn"`
}

type f1jReturn struct {
	Records []alcF1JTrendRecordRef
}

// Axis 1.4 SOAP-encoded arrays can use different child element names.
// Accept any child element and interpret it as an array member.
func (r *f1jReturn) UnmarshalXML(
	d *xml.Decoder,
	start xml.StartElement,
) error {
	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			var record alcF1JTrendRecordRef

			if err := d.DecodeElement(&record, &t); err != nil {
				return err
			}

			r.Records = append(r.Records, record)

		case xml.EndElement:
			if t.Name == start.Name {
				return nil
			}
		}
	}
}

type alcF1JTrendRecordXML struct {
	SequenceNumber int          `xml:"sequenceNumber"`
	Timestamp      string       `xml:"timestamp"`
	ValueType      F1JValueType `xml:"valueType"`
	RawValue       string       `xml:"rawValue"`
	StatusFlags    int          `xml:"statusFlags"`
	FilterDate     string       `xml:"filterDate"`
}

type F1JValueType int

const (
	F1JValueTypeUnknown         F1JValueType = -1
	F1JValueTypeLogStatus       F1JValueType = 0
	F1JValueTypeBoolean         F1JValueType = 1
	F1JValueTypeReal            F1JValueType = 2
	F1JValueTypeEnum            F1JValueType = 3
	F1JValueTypeUnsigned        F1JValueType = 4
	F1JValueTypeInteger         F1JValueType = 5
	F1JValueTypeBitString       F1JValueType = 6
	F1JValueTypeNull            F1JValueType = 7
	F1JValueTypeFailure         F1JValueType = 8
	F1JValueTypeTimeChange      F1JValueType = 9
	F1JValueTypeAny             F1JValueType = 10
	F1JValueTypeAlcStatus       F1JValueType = 11
	F1JValueTypeEvent           F1JValueType = 12
	F1JValueTypeList            F1JValueType = 13
	F1JValueTypeReservedComplex F1JValueType = 14
	F1JValueTypeDouble          F1JValueType = 15
)

func (t F1JValueType) String() string {
	switch t {
	case F1JValueTypeUnknown:
		return "Unknown"
	case F1JValueTypeLogStatus:
		return "LogStatus"
	case F1JValueTypeBoolean:
		return "Boolean"
	case F1JValueTypeReal:
		return "Real"
	case F1JValueTypeEnum:
		return "Enum"
	case F1JValueTypeUnsigned:
		return "Unsigned"
	case F1JValueTypeInteger:
		return "Integer"
	case F1JValueTypeBitString:
		return "BitString"
	case F1JValueTypeNull:
		return "Null"
	case F1JValueTypeFailure:
		return "Failure"
	case F1JValueTypeTimeChange:
		return "TimeChange"
	case F1JValueTypeAny:
		return "Any"
	case F1JValueTypeAlcStatus:
		return "AlcStatus"
	case F1JValueTypeEvent:
		return "Event"
	case F1JValueTypeList:
		return "List"
	case F1JValueTypeReservedComplex:
		return "ReservedComplex"
	case F1JValueTypeDouble:
		return "Double"
	default:
		return fmt.Sprintf("Unknown(%d)", int(t))
	}
}

// Axis may return:
//
//	<item href="#id0"/>
//
// with:
//
//	<multiRef id="id0">...</multiRef>
//
// or it may put the record directly in the array.
type alcF1JTrendRecordRef struct {
	Href string
	Raw  alcF1JTrendRecordXML
}

func (r *alcF1JTrendRecordRef) UnmarshalXML(
	d *xml.Decoder,
	start xml.StartElement,
) error {
	for _, attr := range start.Attr {
		if attr.Name.Local == "href" {
			r.Href = strings.TrimPrefix(attr.Value, "#")
			break
		}
	}

	return d.DecodeElement(&r.Raw, &start)
}

type alcF1JMultiRef struct {
	ID  string
	Raw alcF1JTrendRecordXML
}

func (r *alcF1JMultiRef) UnmarshalXML(
	d *xml.Decoder,
	start xml.StartElement,
) error {
	for _, attr := range start.Attr {
		if attr.Name.Local == "id" {
			r.ID = strings.TrimPrefix(attr.Value, "#")
			break
		}
	}

	return d.DecodeElement(&r.Raw, &start)
}

// F1JTrendRecord is the useful portion of WebCTRL's CoreTrendRecord.
type F1JTrendRecord struct {
	SequenceNumber int
	Timestamp      time.Time
	ValueType      F1JValueType
	RawValue       string
	StatusFlags    int
	FilterDate     *time.Time
}

func parseAxisDateTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)

	if value == "" {
		return time.Time{}, nil
	}

	//
	// Be tolerant of Axis returning milliseconds rather than an
	// xsd:dateTime representation.
	//
	if millis, err := strconv.ParseInt(value, 10, 64); err == nil {
		return time.UnixMilli(millis), nil
	}

	layouts := []string{
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05.999Z07:00",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, value); err == nil {
			return t, nil
		}
	}

	return time.Time{},
		fmt.Errorf("unable to parse Axis dateTime %q", value)
}

func newF1JTrendRecord(
	raw alcF1JTrendRecordXML,
) (F1JTrendRecord, error) {
	timestamp, err := parseAxisDateTime(raw.Timestamp)
	if err != nil {
		return F1JTrendRecord{}, err
	}

	var filterDate *time.Time

	if strings.TrimSpace(raw.FilterDate) != "" {
		t, err := parseAxisDateTime(raw.FilterDate)
		if err != nil {
			return F1JTrendRecord{}, err
		}

		filterDate = &t
	}

	return F1JTrendRecord{
		SequenceNumber: raw.SequenceNumber,
		Timestamp:      timestamp,
		ValueType:      F1JValueType(raw.ValueType),
		RawValue:       raw.RawValue,
		StatusFlags:    raw.StatusFlags,
		FilterDate:     filterDate,
	}, nil
}

// GetF1JTrendData performs one SOAP request.
//
// This is the low-level F1J-specific API.  The normal GetTrendData()
// API below handles paging and converts the result into TrendEvents.
func (s *F1JTrendService) GetF1JTrendData(
	gql string,
	startTime time.Time,
	endTime time.Time,
	fromStart bool,
	maxRecords int,
) ([]F1JTrendRecord, error) {

	request := alcEnvelope{
		XMLNsSoap: "http://schemas.xmlsoap.org/soap/envelope/",
		XMLNsAlc:  "http://soap.core.green.controlj.com",
		Body: alcBody{
			Payload: alcGetF1JTrendDataRequest{
				User:           s.User,
				Passwd:         s.password,
				Path:           gql,
				Start:          startTime.UnixMilli(),
				End:            endTime.UnixMilli(),
				LimitFromStart: fromStart,
				MaxRecords:     maxRecords,
			},
		},
	}

	requestPayload, err := xml.Marshal(request)
	if err != nil {
		return nil, err
	}

	response, err := call(
		s.parent.httpClient,
		s.Endpoint,
		s.User,
		s.password,
		xmlheader+string(requestPayload),
	)
	if err != nil {
		return nil, err
	}

	respObj := new(f1jResponseEnvelope)

	if err := xml.Unmarshal(response, respObj); err != nil {
		return nil, err
	}

	if respObj.Body == nil || respObj.Body.Response == nil {
		return nil,
			fmt.Errorf(
				"F1J trend response did not contain getF1JTrendDataResponse",
			)
	}

	if respObj.Body.Response.Return == nil {
		return []F1JTrendRecord{}, nil
	}

	//
	// Build the Axis multiRef lookup table.
	//
	multiRefs := make(
		map[string]alcF1JTrendRecordXML,
		len(respObj.Body.MultiRef),
	)

	for _, ref := range respObj.Body.MultiRef {
		if ref.ID != "" {
			multiRefs[ref.ID] = ref.Raw
		}
	}

	records := make(
		[]F1JTrendRecord,
		0,
		len(respObj.Body.Response.Return.Records),
	)

	for _, item := range respObj.Body.Response.Return.Records {
		raw := item.Raw

		if item.Href != "" {
			resolved, ok := multiRefs[item.Href]
			if !ok {
				return nil,
					fmt.Errorf(
						"F1J trend response references unknown multiRef %q",
						item.Href,
					)
			}

			raw = resolved
		}

		record, err := newF1JTrendRecord(raw)
		if err != nil {
			return nil, err
		}

		records = append(records, record)
	}

	return records, nil
}

// f1jTrendPoint converts the richer CoreTrendRecord representation
// into the existing TrendPoint representation used by the rest of
// this package.
//
// Returning false means this record is not a numeric trend sample.
// Unlike the old SOAP endpoint, F1J actually lets us distinguish
// these records.
func f1jTrendPoint(record F1JTrendRecord) (TrendPoint, bool) {
	valueString := strings.TrimSpace(record.RawValue)

	value, err := strconv.ParseFloat(valueString, 64)
	if err != nil {
		return TrendPoint{}, false
	}

	return TrendPoint{
		Time:        record.Timestamp,
		TimeString:  TimeToAlcFormat(record.Timestamp),
		ValueString: valueString,
		Value:       value,
	}, true
}

// getTrendDataChunk gives F1JTrendService the same internal interface
// as TrendService.
//
// This is also what TrendGrouper.Preload ultimately wants.
func (s *F1JTrendService) getTrendDataChunk(
	gql string,
	startTime time.Time,
	endTime time.Time,
	fromStart bool,
	maxRecords int,
) ([]TrendPoint, error) {

	records, err := s.GetF1JTrendData(
		gql,
		startTime,
		endTime,
		fromStart,
		maxRecords,
	)
	if err != nil {
		return nil, err
	}

	points := make([]TrendPoint, 0, len(records))

	for _, record := range records {
		point, ok := f1jTrendPoint(record)
		if !ok {
			continue
		}

		points = append(points, point)
	}

	return points, nil
}

// Compatibility layer for grouped trends
func (s *F1JTrendService) GetTrendData(
	gql string,
	startTime time.Time,
	endTime time.Time,
) <-chan TrendEvent {

	output := make(chan TrendEvent)

	go func() {
		defer close(output)

		records, err := s.GetF1JTrendData(
			gql,
			startTime,
			endTime,
			true,
			s.chunkSize,
		)

		source := TrendSource{
			Server:   s.parent.Name,
			Location: gql,
		}

		if err != nil {
			output <- TrendEvent{
				Source: source,
				Err:    err,
			}
			return
		}

		for _, record := range records {
			value, err := strconv.ParseFloat(record.RawValue, 64)
			if err != nil {
				//
				// Not a numeric trend sample -- status record,
				// time change, etc.
				//
				continue
			}

			output <- TrendEvent{
				Source: source,
				Data: TrendPoint{
					Time:        record.Timestamp,
					TimeString:  TimeToAlcFormat(record.Timestamp),
					ValueString: record.RawValue,
					Value:       value,
				},
			}
		}
	}()

	return output
}

func (s *F1JTrendService) GetMultipleTrends(
	startTime time.Time,
	stopTime time.Time,
	locations []string,
) <-chan TrendEvent {

	merger := CreateTrendMerger()
	channels := make([]<-chan TrendEvent, len(locations))

	for i, location := range locations {
		channels[i] = s.GetTrendData(
			location,
			startTime,
			stopTime,
		)
	}

	return merger.Merge(channels)
}

func (s *F1JTrendService) MakeTrendSource(
	server string,
	location string,
) TrendSource {
	return TrendSource{
		Server:   server,
		Location: location,
	}
}
