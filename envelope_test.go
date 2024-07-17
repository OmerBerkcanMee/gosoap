package soap

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/OmerBerkcanMee/xml"
)

var envelopeName = xml.Name{
	Space: soapEnvNS,
	Local: "Envelope",
}
var bodyName = xml.Name{
	Space: soapEnvNS,
	Local: "Body",
}

type headerExample struct {
	XMLName xml.Name `xml:"ns HeaderExample"`
	Attr1   int32    `xml:"attr1,attr"`
	Value   string   `xml:",chardata"`
}

type envelopeExampleField struct {
	XMLName xml.Name `xml:"ContentField"`
	Attr1   string   `xml:"attr1,attr"`
	Attr2   int32    `xml:"attr2,attr"`
	Value   string   `xml:",chardata"`
}

type envelopeContentExample struct {
	XMLName xml.Name             `xml:"ns ContentExample"`
	Attr1   int32                `xml:"attr1,attr"`
	Field1  envelopeExampleField `xml:"ContentField"`
}

type envelopeEncodeTest struct {
	headers    []headerExample
	contentPtr interface{}
	res        string
	err        error
}

var envelopeEncodeTests = []envelopeEncodeTest{
	{
		contentPtr: &envelopeContentExample{
			Attr1: 10,
			Field1: envelopeExampleField{
				XMLName: xml.Name{Local: "ContentField"},
				Attr1:   "test attr",
				Attr2:   11,
				Value:   "This is a test string",
			},
		},
		res: `<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/"><soapenv:Body><ns:ContentExample xmlns:ns="ns" attr1="10"><ns:ContentField attr1="test attr" attr2="11">This is a test string</ns:ContentField></ns:ContentExample></soapenv:Body></soapenv:Envelope>`,
	},
	{
		contentPtr: &envelopeContentExample{
			XMLName: xml.Name{Local: "ContentExample"},
			Attr1:   10,
			Field1: envelopeExampleField{
				XMLName: xml.Name{Local: "ContentField"},
				Attr1:   "test attr",
				Attr2:   11,
				Value:   "This is a test string",
			},
		},
		headers: []headerExample{
			{
				Attr1: 15,
				Value: "test header value",
			},
		},
		res: `<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/"><soapenv:Header><ns:HeaderExample xmlns:ns="ns" attr1="15">test header value</ns:HeaderExample></soapenv:Header><soapenv:Body><ns:ContentExample xmlns:ns="ns" attr1="10"><ns:ContentField attr1="test attr" attr2="11">This is a test string</ns:ContentField></ns:ContentExample></soapenv:Body></soapenv:Envelope>`,
	},
}

func TestEnvelopeEncode(t *testing.T) {
	for i, tt := range envelopeEncodeTests {
		val := NewEnvelope(tt.contentPtr)

		if len(tt.headers) > 0 {
			val.AddHeaders(tt.headers)
		}

		res := new(bytes.Buffer)
		enc := xml.NewEncoder(res)

		if err := enc.Encode(val); !reflect.DeepEqual(err, tt.err) {
			t.Errorf("#%d: %v, want %v", i, err, tt.err)
			continue
		} else if err != nil {
			continue
		}

		if tt.res != res.String() {
			t.Errorf("#%d: mismatch\nhave: `%s`\nwant: `%s`", i, res.String(), tt.res)
			continue
		}
	}
}

type envelopeDecodeTest struct {
	in         string
	contentPtr interface{}
	faultPtr   interface{}
	out        interface{}
	err        error
}

var envelopeDecodeTests = []envelopeDecodeTest{
	{
		in: `<?xml version="1.0"?>
			<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
				<soap:Body>
					<ContentExample xmlns="ns" attr1="10">
						<ContentField attr1="test attr" attr2="11">This is a test content string</ContentField>
					</ContentExample>
				</soap:Body>
			</soap:Envelope>`,
		contentPtr: &envelopeContentExample{},
		out: &Envelope{
			XMLName: envelopeName,
			Body: &Body{
				XMLName: bodyName,
				Content: []any{&envelopeContentExample{
					Attr1: 10,
					Field1: envelopeExampleField{
						Attr1: "test attr",
						Attr2: 11,
						Value: "This is a test content string",
					},
				},
				},
			},
		},
	},
	{
		in: `<?xml version="1.0"?>
			<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
				<soap:Body>
					<soap:Fault>
						<faultcode>FaultCodeValue</faultcode>
						<faultstring>FaultStringValue</faultstring>
						<faultactor>FaultActorValue</faultactor>
						<detail>
							<DetailExample attr1="10">
								<DetailField attr1="test" attr2="11">This is a test string</DetailField>
							</DetailExample>
						</detail>
					</soap:Fault>
				</soap:Body>
			</soap:Envelope>`,
		contentPtr: &envelopeContentExample{},
		faultPtr:   &faultDetailExample{},
		out: &Envelope{
			XMLName: envelopeName,
			Body: &Body{
				XMLName: bodyName,
				Fault: &Fault{
					XMLName: faultName,
					Code:    "FaultCodeValue",
					String:  "FaultStringValue",
					Actor:   "FaultActorValue",
					DetailInternal: &faultDetail{
						Content: `
							<DetailExample attr1="10">
								<DetailField attr1="test" attr2="11">This is a test string</DetailField>
							</DetailExample>
						`,
					},
				},
			},
		},
	},
	{
		in: `<?xml version="1.0"?>
			<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
				<soap:Body>
					<soap:Fault>
						<faultcode>FaultCodeValue</faultcode>
						<faultstring>FaultStringValue</faultstring>
						<faultactor>FaultActorValue</faultactor>
					</soap:Fault>
				</soap:Body>
			</soap:Envelope>`,
		contentPtr: &envelopeContentExample{},
		out: &Envelope{
			XMLName: envelopeName,
			Body: &Body{
				XMLName: bodyName,
				Fault: &Fault{
					XMLName:        faultName,
					Code:           "FaultCodeValue",
					String:         "FaultStringValue",
					Actor:          "FaultActorValue",
					DetailInternal: nil,
				},
			},
		},
	},
	{
		in: `<?xml version="1.0"?>
			<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
				<soap:Body>
					<ContentExample xmlns="ns" attr1="10">
						<ContentField attr1="test attr" attr2="11">This is a test content string</ContentField>
				</soap:Body>
					</ContentExample>
			</soap:Envelope>`,
		contentPtr: &envelopeContentExample{},
		out:        nil,
		err:        &xml.SyntaxError{Msg: "element <ContentExample> closed by </Body>", Line: 6},
	},
	{
		in: `<?xml version="1.0"?>
			<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
				<soap:Body>
					<ContentExample xmlns="ns" attr1="10">
						<ContentField attr1="test attr", attr2="11">This is a test content string</ContentField>
					</ContentExample>
				</soap:Body>
			</soap:Envelope>`,
		contentPtr: &envelopeContentExample{},
		out:        nil,
		err:        &xml.SyntaxError{Msg: "expected attribute name in element", Line: 5},
	},
	{
		in: `<?xml version="1.0"?>
			<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
				<soap:Body>
					<soap:Fault>
						<faultcode>FaultCodeValue</faultcode
						<faultstring>FaultStringValue</faultstring>
						<faultactor>FaultActorValue</faultactor>
						<detail>
							<DetailExample attr1="10">
								<DetailField attr1="test" attr2="11">This is a test string</DetailField>
							</DetailExample>
						</detail>
					</soap:Fault>
				</soap:Body>
			</soap:Envelope>`,
		contentPtr: &envelopeContentExample{},
		out:        nil,
		err:        &xml.SyntaxError{Msg: "invalid characters between </faultcode and >", Line: 6},
	},
	{
		in: `<?xml version="1.0"?>
			<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
				<soap:Body>
					<ContentExample xmlns="ns" attr1="10">
						<ContentField attr1="test attr", attr2="11">This is a test content string</ContentField>
					</ContentExample>
				</soap:Body>
			</soap:Envelope>`,
		contentPtr: nil,
		out:        nil,
		err:        ErrEnvelopeMisconfigured,
	},
}

func TestEnvelopeDecode(t *testing.T) {
	for i, tt := range envelopeDecodeTests {
		var val *Envelope
		val = NewEnvelope(tt.contentPtr)

		dec := xml.NewDecoder(bytes.NewReader([]byte(tt.in)))

		if err := dec.Decode(val); !reflect.DeepEqual(err, tt.err) {
			t.Errorf("#%d: %v, want %v", i, err, tt.err)
			continue
		} else if err != nil {
			continue
		}
		valStr, _ := xml.Marshal(val)
		outStr, _ := xml.Marshal(tt.out)
		if string(valStr) != string(outStr) {
			t.Errorf("#%d: mismatch\nhave: %#+v\nwant: %#+v", i, string(valStr), string(outStr))
			continue
		}
	}
}
