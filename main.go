package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/veggiedefender/protohackers/budgetchat"
	"github.com/veggiedefender/protohackers/means"
	"github.com/veggiedefender/protohackers/primetime"
	"github.com/veggiedefender/protohackers/smoketest"
	"github.com/veggiedefender/protohackers/unusualdatabase"
)

var (
	challengeNum = flag.Int("challenge", -1, "challenge number")
	addr         = flag.String("addr", "0.0.0.0:8080", "listen address")
)

type Challenge interface {
	Listen(addr string)
}

func main() {
	flag.Parse()

	if *challengeNum == -1 {
		fmt.Println("challenge is required")
		os.Exit(1)
	}

	challenges := map[int]Challenge{
		0: smoketest.Server{},
		1: primetime.Server{},
		2: means.Server{},
		3: budgetchat.NewServer(),
		4: unusualdatabase.NewServer(),
	}

	srv, ok := challenges[*challengeNum]
	if !ok {
		log.Fatalf("invalid challenge %d", *challengeNum)
	}

	log.Printf("serving challenge %d on %s", *challengeNum, *addr)
	srv.Listen(*addr)
}
