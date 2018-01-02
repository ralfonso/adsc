package adsc

import "fmt"

type Message interface {
	fmt.Stringer
}

type KeypadMessage struct {
	Fields  *BitFields
	Code    *NumericCode
	Message string
}

type BitFields struct {
	Ready                 bool
	ArmedAway             bool
	ArmedHome             bool
	KeypadBacklight       bool
	KeypadProgrammingMode bool
	BeepNum               int
	ZoneBypassed          bool
	ACPower               bool
	ChimeEnabled          bool
	AlarmOccurred         bool
	AlarmSounding         bool
	BatteryLow            bool
	EntryDelayOff         bool
	Fire                  bool
	SystemIssue           bool
	PerimeterOn           bool
	SysBits               string
	Ademco                bool
	DSC                   bool
}

type NumericCode struct {
	raw       string
	base10val int
}

func (k *KeypadMessage) String() string {
	return k.Message
}

func (n *NumericCode) Value() int {
	return n.base10val
}
