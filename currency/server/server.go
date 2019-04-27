package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"regexp"
	"strings"

	curr "github.com/go-networking/currency/lib0"
)

var currencies = curr.Load("../data.csv")

// This program implements a simple currency lookup service
// over TCP or Unix Data Socket. It loads ISO currency
// information using package lib (see above) and uses a simple
// text-based protocol to interact with the client and send
// the data.
//
// Clients send currency currency search requests as a textual command in the form:
//
// GET <currency, country, or code>
//
// When the server receives the request, it is parsed and is then used
// to seach the list of currencies. The search result is then printed
// line-by-line back to the client.
//
// Focus:
//
// This version of the currency server focuses on implementing a streaming
// strategy when receiving data from client to avoid dripping data when the request is larger than the internal buffer. This relies on the fact that
// net.Conn implements io.Reader which allows the code to stream data.
//
// Testing:
// Netcat or telnet can be used to test this server by connecting and sending command using the format descrived above.
//
// Usage: server0 [options]
// -e host endpoint, default ":4040"

func main() {
	var addr string
	flag.StringVar(&addr, "e", ":4040", "service endpoint")
	flag.Parse()

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("fail to create listener", err)
	}
	defer lis.Close()

	log.Println("*** Global Currency Service ***")
	log.Printf("Service Started: (%s) %s\n", "tcp", addr)

	// connection-loop - handle incoming request
	for {
		conn, err := lis.Accept()
		if err != nil {
			fmt.Println(err)
			if err := conn.Close(); err != nil {
				log.Println("failed to close connection:", err)
			}
			continue
		}
		log.Println("Connected to", conn.RemoteAddr())

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer func() {
		if err := conn.Close(); err != nil {
			log.Println("error closing connection", err)
		}
	}()

	if _, err := conn.Write([]byte("Connected...\nUsage: GET <currency, country, or code>\n")); err != nil {
		log.Println("error writing:", err)
		return
	}

	// appendBytes is a func that simulates end-of-File marker error.
	// Since we will using streaming IO on top a stream protocol, there may never
	// an actual EOF marker. So this function simulates and io.EOF using '\n'
	appendBytes := func(dest, src []byte) ([]byte, error) {
		for _, b := range src {
			if b == '\n' {
				return dest, io.EOF
			}
			dest = append(dest, b)
		}
		return dest, nil
	}

	// loop to stay connected with client unitl client breaks connection
	for {
		var cmdLine []byte

		chunk := make([]byte, 4)
		for {
			n, err := conn.Read(chunk)
			if err != nil {
				if err == io.EOF {
					cmdLine, _ = appendBytes(cmdLine, chunk[:n])
					break
				}
				log.Println("failed to read from client: ", err)
				return
			}
			if cmdLine, err = appendBytes(cmdLine, chunk[:n]); err == io.EOF {
				break
			}
		}

		cmd, param := parseCommandLine(string(cmdLine))
		if cmd == "" {
			if _, err := conn.Write([]byte("Invalid command\n")); err != nil {
				log.Println("Failed to write: ", err)
			}
			continue
		}
		switch strings.ToUpper(cmd) {
		case "GET":
			result := curr.Find(currencies, param)
			if len(result) == 0 {
				if _, err := conn.Write([]byte("Nothing found\n")); err != nil {
					log.Println("Failed to write", err)
					continue
				}
			}
			for _, cur := range result {
				_, err := conn.Write([]byte(fmt.Sprintf("%s %s %s %s\n", cur.Name, cur.Code, cur.Number, cur.Country)))
				if err != nil {
					log.Println("Failed to write", err)
					continue
				}
			}
		default:
			if _, err := conn.Write([]byte("Invalid command\n")); err != nil {
				log.Println("Failed to write: ", err)
				continue
			}
		}
	}
}

func parseCommandLine(cmdLine string) (cmd, param string) {
	r := regexp.MustCompile("'.+'|\".+\"|\\S+")
	parts := r.FindAllString(cmdLine, -1)
	if len(parts) != 2 {
		return "", ""
	}
	cmd = strings.TrimSpace(parts[0])
	param = strings.TrimSpace(parts[1])
	if param[0] == '"' && param[len(param)-1] == '"' {
		param = param[1 : len(param)-1]
	}
	if param[0] == '\'' && param[len(param)-1] == '\'' {
		param = param[1 : len(param)-1]
	}
	return
}
