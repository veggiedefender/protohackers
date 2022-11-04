package speeddaemon

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteMessage(t *testing.T) {
	tests := []struct {
		Message    MessageWriter
		Serialized []byte
	}{
		{
			Message:    MessageError{Msg: "bad"},
			Serialized: []byte{0x10, 0x03, 0x62, 0x61, 0x64},
		},
		{
			Message:    MessageError{Msg: "illegal msg"},
			Serialized: []byte{0x10, 0x0b, 0x69, 0x6c, 0x6c, 0x65, 0x67, 0x61, 0x6c, 0x20, 0x6d, 0x73, 0x67},
		},
		{
			Message: MessageError{Msg: "It is at work everywhere, functioning smoothly at times, at other times in fits and starts. It breathes, it heats, it eats. It shits and fucks. What a mistake to have ever said the id. Everywhere it is machines -- real ones, not figurative ones: machines driving other machines, machines being driven by other machines, with all the necessary couplings and connections."},
			Serialized: func() []byte {
				buf := []byte{0x10, 0xff}
				truncated := "It is at work everywhere, functioning smoothly at times, at other times in fits and starts. It breathes, it heats, it eats. It shits and fucks. What a mistake to have ever said the id. Everywhere it is machines -- real ones, not figurative ones: machin..."
				return append(buf, []byte(truncated)...)
			}(),
		},
		{
			Message: MessageTicket{
				Plate:      "UN1X",
				Road:       66,
				Mile1:      100,
				Timestamp1: 123456,
				Mile2:      110,
				Timestamp2: 123816,
				SpeedX100:  10000,
			},
			Serialized: []byte{0x21, 0x04, 0x55, 0x4e, 0x31, 0x58, 0x00, 0x42, 0x00, 0x64, 0x00, 0x01, 0xe2, 0x40, 0x00, 0x6e, 0x00, 0x01, 0xe3, 0xa8, 0x27, 0x10},
		},
		{
			Message: MessageTicket{
				Plate:      "RE05BKG",
				Road:       368,
				Mile1:      1234,
				Timestamp1: 1000000,
				Mile2:      1235,
				Timestamp2: 1000060,
				SpeedX100:  6000,
			},
			Serialized: []byte{0x21, 0x07, 0x52, 0x45, 0x30, 0x35, 0x42, 0x4b, 0x47, 0x01, 0x70, 0x04, 0xd2, 0x00, 0x0f, 0x42, 0x40, 0x04, 0xd3, 0x00, 0x0f, 0x42, 0x7c, 0x17, 0x70},
		},
		{
			Message:    MessageHeartbeat{},
			Serialized: []byte{0x41},
		},
	}

	for _, test := range tests {
		var buf bytes.Buffer
		bw := bufio.NewWriter(&buf)
		err := WriteMessage(bw, test.Message)
		assert.Nil(t, err)
		assert.Equal(t, test.Serialized, buf.Bytes())
	}
}

func TestReadMessage(t *testing.T) {
	tests := []struct {
		Message      []byte
		Deserialized interface{}
	}{
		{
			Message: []byte{0x20, 0x04, 0x55, 0x4e, 0x31, 0x58, 0x00, 0x00, 0x03, 0xe8},
			Deserialized: MessagePlate{
				Plate:     "UN1X",
				Timestamp: 1000,
			},
		},
		{
			Message: []byte{0x20, 0x07, 0x52, 0x45, 0x30, 0x35, 0x42, 0x4b, 0x47, 0x00, 0x01, 0xe2, 0x40},
			Deserialized: MessagePlate{
				Plate:     "RE05BKG",
				Timestamp: 123456,
			},
		},
		{
			Message:      []byte{0x40, 0x00, 0x00, 0x00, 0x0a},
			Deserialized: MessageWantHeartbeat{Interval: 10},
		},
		{
			Message:      []byte{0x40, 0x00, 0x00, 0x04, 0xdb},
			Deserialized: MessageWantHeartbeat{Interval: 1243},
		},
		{
			Message: []byte{0x80, 0x00, 0x42, 0x00, 0x64, 0x00, 0x3c},
			Deserialized: MessageIAmCamera{
				Road:  66,
				Mile:  100,
				Limit: 60,
			},
		},
		{
			Message: []byte{0x80, 0x01, 0x70, 0x04, 0xd2, 0x00, 0x28},
			Deserialized: MessageIAmCamera{
				Road:  368,
				Mile:  1234,
				Limit: 40,
			},
		},
		{
			Message:      []byte{0x81, 0x01, 0x00, 0x42},
			Deserialized: MessageIAmDispatcher{Roads: []uint16{66}},
		},
		{
			Message:      []byte{0x81, 0x03, 0x00, 0x42, 0x01, 0x70, 0x13, 0x88},
			Deserialized: MessageIAmDispatcher{Roads: []uint16{66, 368, 5000}},
		},
	}

	for _, test := range tests {
		msg, err := ReadMessage(bufio.NewReader(bytes.NewBuffer(test.Message)))
		assert.Nil(t, err)
		assert.Equal(t, test.Deserialized, msg)
	}
}
