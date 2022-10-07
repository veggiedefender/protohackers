package unusualdatabase

import (
	"fmt"
	"log"
	"net"
	"strings"
)

type Server struct {
	KV map[string]string
}

func NewServer() Server {
	return Server{
		KV: make(map[string]string),
	}
}

func (s Server) handleMessage(conn net.PacketConn, sender net.Addr, msg string) error {
	if strings.ContainsRune(msg, '=') {
		split := strings.SplitN(msg, "=", 2)
		s.KV[split[0]] = split[1]
		return nil
	}

	val := s.KV[msg]
	if msg == "version" {
		val = "jesse's cool kv database 1.0"
	}

	_, err := conn.WriteTo([]byte(fmt.Sprintf("%s=%s", msg, val)), sender)
	return err
}

func (s Server) Listen(addr string) {
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	buf := make([]byte, 1000)

	for {
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			log.Println(err)
			continue
		}

		err = s.handleMessage(conn, addr, string(buf[:n]))
		if err != nil {
			log.Println(err)
			continue
		}
	}
}
