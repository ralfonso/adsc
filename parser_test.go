package adsc

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	kpmMsg := &KeypadMessage{}

	tests := []struct {
		name       string
		line       string
		expMessage Message
		expError   error
	}{
		{
			name:       "empty line",
			line:       "",
			expMessage: nil,
			expError:   ErrEmpty,
		},
		{
			name:       "invalid",
			line:       "blah blah blah hey hey",
			expMessage: nil,
			expError:   ErrInvalid,
		},
		{
			name:       "kpm with prefix",
			line:       `!KPM:[00000000000000000000],000,[000000000000000000000000000000],"test message"`,
			expMessage: kpmMsg,
			expError:   nil,
		},
		{
			name:       "kpm without prefix",
			line:       `[00000000000000000000],000,[000000000000000000000000000000],"test message"`,
			expMessage: kpmMsg,
			expError:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			p := NewParser()
			p.parsers["kpm"] = func(l string) (Message, error) {
				return kpmMsg, nil
			}

			msg, err := p.Parse(tt.line)
			if msg != tt.expMessage {
				assert.Equal(tt.expMessage, msg, "unexpected message returned")
			}
			if err != tt.expError {
				assert.Equal(tt.expError, err, "unexpected error returned")
			}
		})
	}
}

func TestKPMParse(t *testing.T) {
	tests := []struct {
		name       string
		line       string
		expMessage Message
		expError   error
	}{
		{
			name: "with prefix",
			line: `!KPM:[00000000000000000A--],000,[000000000000000000000000000000],"test message"`,
			expMessage: &KeypadMessage{
				Fields: &BitFields{
					SysBits: "0",
					Ademco:  true,
				},
				Code: &NumericCode{
					raw: "000",
				},
				Message: "test message",
			},
			expError: nil,
		},
		{
			name: "without prefix",
			line: `[00000000000000000A--],000,[000000000000000000000000000000],"test message"`,
			expMessage: &KeypadMessage{
				Fields: &BitFields{
					SysBits: "0",
					Ademco:  true,
				},
				Code: &NumericCode{
					raw: "000",
				},
				Message: "test message",
			},
			expError: nil,
		},
		{
			name: "bits on",
			line: `!KPM:[1111131111111111BA--],005,[000000000000000000000000000000],"test message"`,
			expMessage: &KeypadMessage{
				Fields: &BitFields{
					Ready:                 true,
					ArmedAway:             true,
					ArmedHome:             true,
					KeypadBacklight:       true,
					KeypadProgrammingMode: true,
					BeepNum:               3,
					ZoneBypassed:          true,
					ACPower:               true,
					ChimeEnabled:          true,
					AlarmOccurred:         true,
					AlarmSounding:         true,
					BatteryLow:            true,
					EntryDelayOff:         true,
					Fire:                  true,
					SystemIssue:           true,
					PerimeterOn:           true,
					SysBits:               "B",
					Ademco:                true,
					DSC:                   false,
				},
				Code: &NumericCode{
					raw:       "005",
					base10val: 5,
				},
				Message: "test message",
			},
			expError: nil,
		},
		{
			name: "fault",
			line: `!KPM:[0000000000000000BA--],005,[000000000000000000000000000000],"  FAULT Pinball Room Door  "`,
			expMessage: &KeypadMessage{
				Fields: &BitFields{
					SysBits: "B",
					Ademco:  true,
				},
				Code: &NumericCode{
					raw:       "005",
					base10val: 5,
				},
				Message: "FAULT Pinball Room Door",
			},
			expError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			msg, err := parseKPM(tt.line)
			assert.Equal(tt.expError, err, "unexpected error returned")
			if tt.expMessage != nil {
				assert.NotNil(msg)
				kMsg, ok := msg.(*KeypadMessage)
				assert.True(ok, "returned message was incorrect type %q", reflect.TypeOf(msg))
				ekMsg := tt.expMessage.(*KeypadMessage)
				assert.Equal(ekMsg.Fields, kMsg.Fields, "unexpected message fields")
				assert.Equal(ekMsg.Code, kMsg.Code, "unexpected message code")
				assert.Equal(ekMsg.Message, kMsg.Message, "unexpected message text")
			}
		})
	}
}
