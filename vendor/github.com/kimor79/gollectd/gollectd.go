// Parse data from the collectd network plugin
// This is modeled after collectd's own network.c
package gollectd

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
)

const (
	TypeCounter  = 0
	TypeGauge    = 1
	TypeDerive   = 2
	TypeAbsolute = 3
)

const (
	// Values taken from commit 633c3966f7 of
	// https://github.com/collectd/collectd/commits/master/src/network.h

	ParseHost           = 0x0000
	ParseTime           = 0x0001
	ParsePlugin         = 0x0002
	ParsePluginInstance = 0x0003
	ParseType           = 0x0004
	ParseTypeInstance   = 0x0005
	ParseValues         = 0x0006
	ParseInterval       = 0x0007
	ParseTimeHR         = 0x0008
	ParseIntervalHR     = 0x0009

	// Notifications
	ParseMessage  = 0x0100
	ParseSeverity = 0x0101

	ParseSignature  = 0x0200
	ParseEncryption = 0x0210
)

var ErrorInvalid = errors.New("Invalid packet")
var ErrorUnsupported = errors.New("Unsupported packet")
var ErrorUnknownType = errors.New("Unknown value type")
var ErrorUnknownDataType = errors.New("Unknown data source type")

var ValueTypeNames = map[string]uint8{
	"absolute": TypeAbsolute,
	"counter":  TypeCounter,
	"derive":   TypeDerive,
	"gauge":    TypeGauge,
}

var ValueTypeValues = map[uint8]string{
	TypeAbsolute: "absolute",
	TypeCounter:  "counter",
	TypeDerive:   "derive",
	TypeGauge:    "gauge",
}

type Packet struct {
	Hostname       string
	Interval       uint64
	IntervalHR     uint64
	Plugin         string
	PluginInstance string
	Time           uint64
	TimeHR         uint64
	Type           string
	TypeInstance   string
	Values         []Value
}

type Type struct {
	Name string
	Type uint8
	Min  string
	Max  string
}

type Types map[string][]*Type

type Value struct {
	Name     string
	Type     uint8
	TypeName string
	Value    float64
}

func (p Packet) FormatName() string {
	metricName := p.Hostname + "/" + p.Plugin

	if len(p.PluginInstance) > 0 {
		metricName += "-" + p.PluginInstance
	}

	metricName += "/" + p.Type

	if len(p.TypeInstance) > 0 {
		metricName += "-" + p.TypeInstance
	}

	return metricName
}

func Packets(b []byte, types Types) (*[]Packet, error) {
	packets := make([]Packet, 0)

	buf := bytes.NewBuffer(b)

	var packetHeader struct {
		PartType   uint16
		PartLength uint16
	}

	var err error
	var packet Packet
	var time uint64
	var valueCount uint16
	var valueTypes []uint8

	for buf.Len() > 0 {
		partType := buf.Next(2)
		if len(partType) < 2 {
			return nil, ErrorInvalid
		}
		packetHeader.PartType = binary.BigEndian.Uint16(partType)

		partLength := buf.Next(2)
		if len(partLength) < 2 {
			return nil, ErrorInvalid
		}
		packetHeader.PartLength = binary.BigEndian.Uint16(partLength)

		if packetHeader.PartLength < 5 {
			return nil, ErrorInvalid
		}

		nextPos := int(packetHeader.PartLength) - 4

		partBytes := buf.Next(nextPos)
		if len(partBytes) < nextPos {
			return nil, ErrorInvalid
		}

		partBuffer := bytes.NewBuffer(partBytes)

		switch packetHeader.PartType {
		case ParseEncryption:
			return nil, ErrorUnsupported
		case ParseHost:
			str := partBuffer.String()
			packet.Hostname = str[0 : len(str)-1]
		case ParseInterval:
			err = binary.Read(partBuffer, binary.BigEndian, &time)
			if err != nil {
				return nil, err
			}

			packet.Interval = time
		case ParseIntervalHR:
			err = binary.Read(partBuffer, binary.BigEndian, &time)
			if err != nil {
				return nil, err
			}

			packet.IntervalHR = time
		case ParseMessage:
			// ignore (notification)
		case ParsePlugin:
			str := partBuffer.String()
			packet.Plugin = str[0 : len(str)-1]
		case ParsePluginInstance:
			str := partBuffer.String()
			packet.PluginInstance = str[0 : len(str)-1]
		case ParseSeverity:
			// ignore (notification)
		case ParseSignature:
			return nil, ErrorUnsupported
		case ParseTime:
			err = binary.Read(partBuffer, binary.BigEndian, &time)
			if err != nil {
				return nil, err
			}

			packet.Time = time
		case ParseTimeHR:
			err = binary.Read(partBuffer, binary.BigEndian, &time)
			if err != nil {
				return nil, err
			}

			packet.TimeHR = time
		case ParseType:
			str := partBuffer.String()
			packet.Type = str[0 : len(str)-1]
		case ParseTypeInstance:
			str := partBuffer.String()
			packet.TypeInstance = str[0 : len(str)-1]
		case ParseValues:
			err = binary.Read(partBuffer, binary.BigEndian, &valueCount)
			if err != nil {
				return nil, err
			}

			valueTypes = make([]uint8, valueCount, valueCount)
			packet.Values = make([]Value, valueCount, valueCount)
			var packetValue Value

			err = binary.Read(partBuffer, binary.BigEndian, &valueTypes)
			if err != nil {
				return nil, err
			}

			for i, t := range valueTypes {
				packetValue.Type = t

				if typeName, ok := ValueTypeValues[t]; ok {
					packetValue.TypeName = typeName
				}

				if _, ok := types[packet.Type]; ok {
					packetValue.Name = types[packet.Type][i].Name
				}

				switch t {
				case TypeAbsolute:
					var value uint64
					err = binary.Read(partBuffer, binary.BigEndian, &value)
					if err != nil {
						return nil, err
					}

					packetValue.Value = float64(value)
				case TypeCounter:
					var value uint64
					err = binary.Read(partBuffer, binary.BigEndian, &value)
					if err != nil {
						return nil, err
					}

					packetValue.Value = float64(value)
				case TypeDerive:
					var value int64
					err = binary.Read(partBuffer, binary.BigEndian, &value)
					if err != nil {
						return nil, err
					}

					packetValue.Value = float64(value)
				case TypeGauge:
					var value float64
					err = binary.Read(partBuffer, binary.LittleEndian, &value)
					if err != nil {
						return nil, err
					}

					packetValue.Value = float64(value)
				default:
					return nil, ErrorUnknownType
				}

				packet.Values[i] = packetValue
			}

			packets = append(packets, packet)
		default:
			// Ignore unknown fields
		}
	}

	return &packets, nil
}

func TypesDBFile(path string) (Types, error) {
	// See https://collectd.org/documentation/manpages/types.db.5.shtml

	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return TypesDB(b)
}

func TypesDB(b []byte) (Types, error) {
	types := make(Types)
	content := string(b)
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty & comment lines
		if line == "" || line[0] == '#' {
			continue
		}
		dataSetName, dataSetSources, err := ParseDataSet(line)
		if err != nil {
			return nil, fmt.Errorf("Line %d: %s", i+1, err.Error())
		}

		types[dataSetName] = dataSetSources
	}

	return types, nil
}

// ParseDataSet parses one line from a collectd types.db file and returns
// the data-set name & a Type struct
func ParseDataSet(s string) (string, []*Type, error) {
	splitFn := func(c rune) bool {
		return c == '\t' || c == ' ' || c == ','
	}
	fields := strings.FieldsFunc(s, splitFn)

	if len(fields) < 2 {
		return "", nil, fmt.Errorf("Minimum of 2 fields required \"%s\"", s)
	}

	dataSetName := fields[0]
	dataSetSources := make([]*Type, len(fields[1:]))

	// Parse each data source
	for i, field := range fields[1:] {
		// Split data source fields
		dsFields := strings.Split(field, ":")
		if len(dsFields) != 4 {
			return "", nil, fmt.Errorf("Exactly 4 fields required \"%s\"", field)
		}

		// Parse data source type
		dsTypeStr := strings.ToLower(dsFields[1])
		dsType, ok := ValueTypeNames[dsTypeStr]
		if !ok {
			return "", nil, fmt.Errorf("Invalid data-source type \"%s\"", dsTypeStr)
		}

		dataSetSources[i] = &Type{
			Name: dsFields[0],
			Type: dsType,
			Min:  dsFields[2],
			Max:  dsFields[3],
		}
	}

	return dataSetName, dataSetSources, nil
}
