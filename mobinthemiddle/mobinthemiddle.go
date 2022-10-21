package mobinthemiddle

import (
	"bufio"
	"log"
	"net"
	"regexp"
)

type Server struct{}

var BogusCoinAddress = regexp.MustCompile(`(\b)7[a-zA-Z0-9_]{25,34}(\n| )`)

func (s *Server) Listen(addr string) {
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

func handleConnection(eyeball net.Conn) {
	defer eyeball.Close()

	origin, err := net.Dial("tcp", "chat.protohackers.com:16963")
	if err != nil {
		return
	}
	defer origin.Close()

	eyeballChan := chanFromConn(eyeball)
	originChan := chanFromConn(origin)

	for {
		select {
		case msg := <-eyeballChan:
			if msg == nil {
				return
			}
			if err := intercept(origin, msg); err != nil {
				log.Println(err)
				return
			}
		case msg := <-originChan:
			if msg == nil {
				return
			}
			if err := intercept(eyeball, msg); err != nil {
				log.Println(err)
				return
			}
		}
	}
}

func intercept(recipient net.Conn, msg []byte) error {
	msg = BogusCoinAddress.ReplaceAll(msg, []byte("${1}7YWHMfk9JZe0LM0g1ZauHuiSxhI${2}"))

	_, err := recipient.Write(msg)
	if err != nil {
		return err
	}

	return nil
}

func chanFromConn(conn net.Conn) chan []byte {
	c := make(chan []byte)

	go func() {
		reader := bufio.NewReader(conn)
		defer close(c)
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				return
			}

			c <- line
		}
	}()

	return c
}
