package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/go-rod/bartender"
)

func main() {
	port := flag.String("p", ":3001", "port to listen on")
	target := flag.String("t", "", "target url to proxy")

	flag.Parse()

	if *target == "" {
		panic("cli option -t required")
	}

	log.Printf("Bartender started %s -> %s\n", *port, *target)

	err := http.ListenAndServe(*port, bartender.New(*port, *target))
	if err != nil {
		log.Fatalln(err)
	}
}
