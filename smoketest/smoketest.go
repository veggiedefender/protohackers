package smoketest

import (
	"io"
	"log"
	"net"
)

type Server struct{}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	io.Copy(conn, conn)
}

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
