package main

import (
	"flag"
)

var (
	url, file *string
)

func init() {
	url = flag.String("m", "", "Marathon URL")
	file = flag.String("f", "", "Job file")
}

func main() {
	flag.Parse()
}
