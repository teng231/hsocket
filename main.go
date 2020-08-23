package main

import (
	"flag"
)

var port = ""

func init() {
	flag.StringVar(&port, "port", ":8000", "http port")

}

func main() {
	flag.Parse()
	start()
}
