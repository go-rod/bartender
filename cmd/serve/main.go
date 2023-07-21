package main

import (
	"flag"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-rod/bartender"
)

func main() {
	port := flag.String("p", ":3001", "port to listen on")
	target := flag.String("t", "", "target url to proxy")
	size := flag.Int("s", 2, "size of the pool")
	maxWait := flag.Duration("w", 3*time.Second, "max wait time for a page rendering")

	var block BlockRequestsFlag
	flag.Var(&block, "b", "block the requests that match the pattern, such as 'https://a.com/*', can set multiple ones")

	flag.Parse()

	if *target == "" {
		panic("cli option -t required")
	}

	log.Printf("Bartender started %s -> %s\n", *port, *target)

	b := bartender.New(*port, *target, *size)
	b.BlockRequest(block...)
	b.MaxWait(*maxWait)
	b.WarnUp()

	err := http.ListenAndServe(*port, b)
	if err != nil {
		log.Fatalln(err)
	}
}

type BlockRequestsFlag []string

func (i *BlockRequestsFlag) String() string {
	return strings.Join(*i, ", ")
}

func (i *BlockRequestsFlag) Set(value string) error {
	*i = append(*i, value)

	return nil
}
