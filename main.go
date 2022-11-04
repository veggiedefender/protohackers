package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/veggiedefender/protohackers/budgetchat"
	"github.com/veggiedefender/protohackers/means"
	"github.com/veggiedefender/protohackers/mobinthemiddle"
	"github.com/veggiedefender/protohackers/primetime"
	"github.com/veggiedefender/protohackers/smoketest"
	"github.com/veggiedefender/protohackers/speeddaemon"
	"github.com/veggiedefender/protohackers/unusualdatabase"
	"golang.org/x/exp/maps"
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

	challenges := map[int]Challenge{
		0: smoketest.Server{},
		1: primetime.Server{},
		2: means.Server{},
		3: budgetchat.NewServer(),
		4: unusualdatabase.NewServer(),
		5: mobinthemiddle.Server{},
		6: speeddaemon.Server{},
	}

	if *challengeNum == -1 {
		implemented := maps.Keys(challenges)
		sort.Ints(implemented)
		fmt.Printf("challenge is required: %v\n", implemented)
		os.Exit(1)
	}

	srv, ok := challenges[*challengeNum]
	if !ok {
		log.Fatalf("invalid challenge %d", *challengeNum)
	}

	log.Printf("serving challenge %d on %s", *challengeNum, *addr)
	srv.Listen(*addr)
}
