package speeddaemon

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"math"
)

const (
	TypeError         byte = 0x10
	TypePlate         byte = 0x20
	TypeTicket        byte = 0x21
	TypeWantHeartbeat byte = 0x40
	TypeHeartbeat     byte = 0x41
	TypeIAmCamera     byte = 0x80
	TypeIAmDispatcher byte = 0x81
)

var (
	ErrNotImplemented = errors.New("message not implemented")
)

type MessageError struct {
	Msg string
}

type MessagePlate struct {
	Plate     string
	Timestamp uint32
}

type MessageTicket struct {
	Plate      string
	Road       uint16
	Mile1      uint16
	Timestamp1 uint32
	Mile2      uint16
	Timestamp2 uint32
	SpeedX100  uint16
}

type MessageWantHeartbeat struct {
	Interval uint32 // deciseconds
}

type MessageHeartbeat struct{}

type MessageIAmCamera struct {
	Road  uint16
	Mile  uint16
	Limit uint16
}

type MessageIAmDispatcher struct {
	Roads []uint16
}

type MessageWriter interface {
	Write(w *bufio.Writer)
}

func ReadMessage(r *bufio.Reader) (interface{}, error) {
	typ, err := r.ReadByte()
	if err != nil {
		return nil, err
	}

	er := errorReader{r: r}

	switch typ {
	case TypePlate:
		msg := MessagePlate{
			Plate:     er.readString(),
			Timestamp: er.readUint32(),
		}
		return msg, er.err

	case TypeWantHeartbeat:
		msg := MessageWantHeartbeat{
			Interval: er.readUint32(),
		}
		return msg, er.err

	case TypeIAmCamera:
		msg := MessageIAmCamera{
			Road:  er.readUint16(),
			Mile:  er.readUint16(),
			Limit: er.readUint16(),
		}
		return msg, er.err

	case TypeIAmDispatcher:
		numRoads, err := r.ReadByte()
		if err != nil {
			return nil, err
		}

		roads := make([]uint16, 0, numRoads)
		for i := 0; i < int(numRoads); i++ {
			roads = append(roads, er.readUint16())
		}
		if er.err != nil {
			return nil, err
		}

		return MessageIAmDispatcher{Roads: roads}, nil

	// Server->Client types unimplemented
	case TypeError:
		return MessageError{}, nil
	case TypeTicket:
		return MessageTicket{}, nil
	case TypeHeartbeat:
		return MessageHeartbeat{}, nil

	default:
		return nil, ErrNotImplemented
	}
}

func WriteMessage(w *bufio.Writer, msg MessageWriter) error {
	msg.Write(w)
	return w.Flush()
}

func (e MessageError) Write(w *bufio.Writer) {
	writeByte(w, TypeError)
	writeString(w, e.Msg)
}

func (t MessageTicket) Write(w *bufio.Writer) {
	writeByte(w, TypeTicket)
	writeString(w, t.Plate)
	writeUint16(w, t.Road)
	writeUint16(w, t.Mile1)
	writeUint32(w, t.Timestamp1)
	writeUint16(w, t.Mile2)
	writeUint32(w, t.Timestamp2)
	writeUint16(w, t.SpeedX100)
}

func (h MessageHeartbeat) Write(w *bufio.Writer) {
	writeByte(w, TypeHeartbeat)
}

type errorReader struct {
	r   *bufio.Reader
	err error
}

func (er *errorReader) readString() string {
	if er.err != nil {
		return ""
	}

	n, err := er.r.ReadByte()
	if err != nil {
		er.err = err
		return ""
	}

	buf := make([]byte, n)
	_, err = io.ReadFull(er.r, buf)
	if err != nil {
		er.err = err
		return ""
	}

	return string(buf)
}

func (er *errorReader) readUint16() uint16 {
	if er.err != nil {
		return 0
	}
	var i uint16
	er.err = binary.Read(er.r, binary.BigEndian, &i)
	return i
}

func (e *errorReader) readUint32() uint32 {
	if e.err != nil {
		return 0
	}
	var i uint32
	e.err = binary.Read(e.r, binary.BigEndian, &i)
	return i
}

func writeByte(w *bufio.Writer, b byte) {
	w.WriteByte(b)
}

func writeString(w *bufio.Writer, s string) {
	if len(s) > math.MaxUint8 {
		s = s[:math.MaxUint8-3] + "..."
	}
	w.WriteByte(byte(len(s)))
	w.WriteString(s)
}

func writeUint16(w *bufio.Writer, i uint16) {
	var buf [2]byte
	binary.BigEndian.PutUint16(buf[:], i)
	w.Write(buf[:])
}

func writeUint32(w *bufio.Writer, i uint32) {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], i)
	w.Write(buf[:])
}
