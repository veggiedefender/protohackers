package means

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
)

type Insert struct {
	Timestamp int32
	Price     int32
}

type Query struct {
	MinTime int32
	MaxTime int32
}

func readMessage(r io.Reader) (interface{}, error) {
	var buf [9]byte
	_, err := io.ReadFull(r, buf[:])
	if err != nil {
		return nil, err
	}

	typ := buf[0]
	first := int32(binary.BigEndian.Uint32(buf[1:5]))
	second := int32(binary.BigEndian.Uint32(buf[5:9]))

	switch typ {
	case 'I':
		return Insert{first, second}, nil
	case 'Q':
		return Query{first, second}, nil
	default:
		return nil, fmt.Errorf("invalid type: %c", typ)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	prices := make([]Insert, 0)

	for {
		msg, err := readMessage(conn)
		if err != nil {
			return
		}

		switch m := msg.(type) {
		case Insert:
			prices = append(prices, m)
		case Query:
			sum := 0
			count := 0

			for _, price := range prices {
				if m.MinTime <= price.Timestamp && price.Timestamp <= m.MaxTime {
					sum += int(price.Price)
					count += 1
				}
			}

			average := 0
			if count > 0 {
				average = sum / count
			}

			binary.Write(conn, binary.BigEndian, int32(average))
		default:
			panic(fmt.Sprintf("unexpected message type: %T", m))
		}
	}
}

type Server struct{}

func (s Server) Listen(addr string) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("tcp server accept error", err)
		}

		go handleConnection(conn)
	}
}
