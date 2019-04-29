package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	curr "github.com/go-networking/currency/lib"
)

func main() {
	var addr string
	flag.StringVar(&addr, "e", "localhost:4040", "service endpoint")
	flag.Parse()

	dialer := &net.Dialer{
		Timeout:   time.Second * 300,
		KeepAlive: time.Minute * 5,
	}

	var (
		conn           net.Conn
		err            error
		connTries      = 0
		connMaxRetries = 3
		connSleepRetry = time.Second * 1
	)

	for connTries < connMaxRetries {
		fmt.Println("creating connection socket to", addr)
		conn, err = dialer.Dial("tcp", addr)
		if err != nil {
			fmt.Println("failed to create socket to", addr)
			switch err := err.(type) {
			case net.Error:
				if err.Temporary() {
					connTries++
					fmt.Println("trying again in:", connSleepRetry)
					time.Sleep(connSleepRetry)
					continue
				}

				log.Fatal("unable to recover")
			default:
				os.Exit(1)
			}
		}
		break
	}

	if conn == nil {
		fmt.Println("failed to create a connection successfully")
		os.Exit(1)
	}
	defer conn.Close()
	fmt.Println("connected to currency service: ", addr)

	var param string

	for {
		fmt.Print("currency >")
		_, err = fmt.Scanf("%s", &param)
		if err != nil {
			fmt.Println("Usage: <search string or *>")
			continue
		}

		req := curr.CurrencyRequest{Get: param}

		if err := json.NewEncoder(conn).Encode(&req); err != nil {
			switch err := err.(type) {
			case net.Error:
				log.Fatal("failed to send request:", err)
			default:
				fmt.Println("failed to encode request:", err)
				continue
			}
		}

		var currencies []curr.Currency
		err = json.NewDecoder(conn).Decode(&currencies)
		if err != nil {
			switch err := err.(type) {
			case net.Error:
				log.Fatal("failed to receive response:", err)
			default:
				fmt.Println("failed to decode response:", err)
				continue
			}
		}
		fmt.Println(currencies)
	}
}
