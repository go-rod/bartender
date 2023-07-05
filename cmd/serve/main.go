package main

import (
	"flag"
	"fmt"

	"github.com/go-rod/bartender"
)

func main() {
	port := flag.String("p", ":3001", "port to listen on")
	target := flag.String("t", "", "target url to proxy")

	flag.Parse()

	if *target == "" {
		panic("cli option -t required")
	}

	fmt.Printf("Bartender started %s -> %s\n", *port, *target)

	bartender.New(*port, *target).Serve()
}
