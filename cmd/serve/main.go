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

	var bypassUAs StringsFlag = bartender.DefaultBypassUserAgentNames
	flag.Var(&bypassUAs, "u", "bypass the specified user-agent names")

	var blockList StringsFlag
	flag.Var(&blockList, "b",
		"block the requests that match the pattern, such as 'https://a.com/*', can set multiple ones")

	flag.Parse()

	if *target == "" {
		panic("cli option -t required")
	}

	log.Printf("Bartender started %s -> %s\n", *port, *target)
	log.Printf("Block list: %v\n", blockList)
	log.Printf("Bypass user-agent names: %v\n", bypassUAs)

	b := bartender.New(*port, *target, *size)
	b.BlockRequests(blockList...)
	b.BypassUserAgentNames(bypassUAs...)
	b.MaxWait(*maxWait)
	b.WarmUp()
	b.AutoFree()

	err := http.ListenAndServe(*port, b)
	if err != nil {
		log.Fatalln(err)
	}
}

type StringsFlag []string

func (i *StringsFlag) String() string {
	return strings.Join(*i, ", ")
}

func (i *StringsFlag) Set(value string) error {
	*i = append(*i, value)

	return nil
}
