package adsc

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

var (
	ErrEmpty     = errors.New("empty line")
	ErrInvalid   = errors.New("invalid message")
	ErrUnhandled = errors.New("unhandled message")
)

type parserFn func(string) (Message, error)

var (
	// !KPM:[10000001100000003A--],008,[f70600051008001c28020000000000]," DISARMED CHIME   Ready to Arm  "
	kpmRE = regexp.MustCompile(`(?:!KPM:)?\[([0-9A-F\-]{20})\],([0-9A-F]{3}),\[([0-9a-f]{30})\],"(.+?)"`)

	defaultParsers = map[string]parserFn{
		"kpm": parseKPM,
	}
)

type Parser struct {
	parsers map[string]parserFn
}

func NewParser() *Parser {
	return &Parser{
		parsers: defaultParsers,
	}
}

func (p *Parser) Parse(line string) (Message, error) {
	if len(line) == 0 {
		return nil, ErrEmpty
	}

	if kpmRE.MatchString(line) {
		return p.getParser("kpm", line)
	}

	if strings.HasPrefix(line, "!AUI") {
		return p.getParser("aui", line)
	}

	if strings.HasPrefix(line, "!RFX") {
		return p.getParser("rfx", line)
	}

	if strings.HasPrefix(line, "!SER2SOCK") {
		return p.getParser("ser2sock", line)
	}

	return nil, ErrInvalid
}

func (p *Parser) getParser(commandType, line string) (Message, error) {
	pp, ok := p.parsers[commandType]
	if !ok {
		return nil, ErrUnhandled
	}

	return pp(line)
}

func parseKPM(line string) (Message, error) {
	result := kpmRE.FindStringSubmatch(line)
	if result == nil || len(result) != 5 {
		return nil, ErrInvalid
	}

	bitFields, err := parseBitFields(result[1])
	if err != nil {
		return nil, err
	}

	ncode, err := parseNumericCode(result[2])
	if err != nil {
		return nil, err
	}

	return &KeypadMessage{
		Fields:  bitFields,
		Code:    ncode,
		Message: strings.TrimSpace(result[4]),
	}, nil
}

func parseBitFields(bitField string) (*BitFields, error) {
	var err error
	bf := &BitFields{}
	if bitField[0] == '1' {
		bf.Ready = true
	}
	if bitField[1] == '1' {
		bf.ArmedAway = true
	}
	if bitField[2] == '1' {
		bf.ArmedHome = true
	}
	if bitField[3] == '1' {
		bf.KeypadBacklight = true
	}
	if bitField[4] == '1' {
		bf.KeypadProgrammingMode = true
	}

	bf.BeepNum, err = strconv.Atoi(string(bitField[5]))
	if err != nil {
		return nil, err
	}

	if bitField[6] == '1' {
		bf.ZoneBypassed = true
	}
	if bitField[7] == '1' {
		bf.ACPower = true
	}
	if bitField[8] == '1' {
		bf.ChimeEnabled = true
	}
	if bitField[9] == '1' {
		bf.AlarmOccurred = true
	}
	if bitField[10] == '1' {
		bf.AlarmSounding = true
	}
	if bitField[11] == '1' {
		bf.BatteryLow = true
	}
	if bitField[12] == '1' {
		bf.EntryDelayOff = true
	}
	if bitField[13] == '1' {
		bf.Fire = true
	}
	if bitField[14] == '1' {
		bf.SystemIssue = true
	}
	if bitField[15] == '1' {
		bf.PerimeterOn = true
	}

	bf.SysBits = string(bitField[16])

	if bitField[17] == 'A' {
		bf.Ademco = true
	}
	if bitField[17] == 'D' {
		bf.DSC = true
	}

	return bf, nil
}

func parseNumericCode(raw string) (*NumericCode, error) {
	val, err := strconv.ParseInt(raw, 10, 0)
	if err != nil {
		return nil, err
	}

	return &NumericCode{
		raw:       raw,
		base10val: int(val),
	}, nil
}
