package main

import (
	"encoding/json"
	"flag"
	"log"
	"net"

	curr "github.com/go-networking/currency/lib"
)

var currencies = curr.Load("../data.csv")

// This program implements a simple currency lookup service
// over TCP or Unix Data Socket. It loads ISO currency
// informatio using package curr (see above) and uses a simple
// JSON-encode text-based protocol to exchange data with a client.
//
// Clients send currency search requests as JSON objects
// as {"Get":"<currency name, code, or country"}. The request data is
// then unmarshalled to Go type curr.CurrencyRequest using
// the encoding/json package.
//
// The request is then used to seach the list of
// currencies. The search result, a []curr.Currency, is marshalled
// as JSON array of objects and sent to the client.
//
// Focus:
// This version of the program highlights the use of the encoding
// packages to serialize data to/from Go data types to another
// representation such as JSON. This version uses the encoding/json
// package Encoder/Decoder types which are accept an io.Writer and
// io.Reader respectively. This means they can be used directly with
// the io.Conn value
//
// Testing:
// Netcat can be used for rudimentary testing.
//
// Usage: server [options]
// options:
//  -e host endpoint, default ":4040"

func main() {
	var addr string
	flag.StringVar(&addr, "e", ":4040", "host endpoint")
	flag.Parse()

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("Can not create listener", err)
	}
	defer lis.Close()
	log.Println("*** Global Currency Service ***")
	log.Printf("Service started: (tcp) %s\n", addr)

	// connection loop
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println(err)
			conn.Close()
			continue
		}
		log.Println("Connected to", conn.RemoteAddr())
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer func() {
		if err := conn.Close(); err != nil {
			log.Println(err)
		}
	}()

	// create json encoder/decoder using net.Conn as
	// io.Writer and io.Reader for streaming IO
	dec := json.NewDecoder(conn)
	enc := json.NewEncoder(conn)

	for {
		var req curr.CurrencyRequest
		if err := dec.Decode(&req); err != nil {
			log.Println("failed to decode request:", err)
			return
		}

		results := curr.Find(currencies, req.Get)

		if err := enc.Encode(&results); err != nil {
			log.Println("failed to encode data:", err)
			return
		}
	}
}
