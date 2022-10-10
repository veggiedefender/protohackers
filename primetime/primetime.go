package primetime

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net"
)

type Request struct {
	Method string   `json:"method"`
	Number *float64 `json:"number"`
}

type Response struct {
	Method  string `json:"method"`
	IsPrime bool   `json:"prime"`
}

func parseRequest(req Request) (float64, error) {
	if req.Method != "isPrime" {
		return 0, fmt.Errorf("invalid method: %q", req.Method)
	}

	if req.Number == nil {
		return 0, errors.New("number is missing")
	}

	return *req.Number, nil
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	encoder := json.NewEncoder(conn)

	for scanner.Scan() {
		var req Request

		err := json.Unmarshal(scanner.Bytes(), &req)
		if err != nil {
			encoder.Encode(Response{})
			return
		}

		n, err := parseRequest(req)
		if err != nil {
			encoder.Encode(Response{})
			return
		}

		response := Response{
			Method:  "isPrime",
			IsPrime: big.NewInt(int64(n)).ProbablyPrime(0),
		}

		err = encoder.Encode(response)
		if err != nil {
			return
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
