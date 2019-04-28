package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"time"

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

	tries := 0
	delayTime := time.Millisecond * 10

	// connection loop
	for {
		conn, err := lis.Accept()

		if err != nil {
			switch e := err.(type) {
			case net.Error:
				if e.Temporary() {
					if tries > 5 {
						conn.Close()
						log.Fatalf("Can not establish connection after %d times: %v\n", tries, err)
					}
					tries++
					delayTime *= 2
					time.Sleep(delayTime)
				} else {
					conn.Close()
					continue
				}
			default:
				log.Println(err)
				conn.Close()
				continue
			}
			tries = 0
			delayTime = time.Millisecond * 10
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

	for {
		// set initial deadline prior to entering
		// the client request/response loop to 45 seconds.
		// This means that the client has 45 seconds to send
		// its initial request or loose the connection.
		if err := conn.SetDeadline(time.Now().Add(time.Second * 45)); err != nil {
			log.Println("Failed to set deadline", err)
			return
		}

		var req curr.CurrencyRequest
		dec := json.NewDecoder(conn)
		if err := dec.Decode(&req); err != nil {
			switch err := err.(type) {
			case net.Error:
				if err.Timeout() {
					log.Println("deadline reached, disconnecting...")
					return
				}
				fmt.Println("network error:", err)
			default:
				if err == io.EOF {
					log.Println("closing connection:", err)
					return
				}
				enc := json.NewEncoder(conn)
				if encerr := enc.Encode(&curr.CurrencyError{Error: err.Error()}); encerr != nil {
					log.Println("Failed error encoding:", encerr)
					return
				}
				continue
			}
		}

		results := curr.Find(currencies, req.Get)

		enc := json.NewEncoder(conn)
		if err := enc.Encode(&results); err != nil {
			switch err := err.(type) {
			case net.Error:
				log.Println("failed to send response:", err)
				return
			default:
				if encerr := enc.Encode(&curr.CurrencyError{Error: err.Error()}); encerr != nil {
					log.Println("failed to send error:", encerr)
					return
				}
				continue
			}
		}
	}
}
