package main

import (
	"flag"
	"log"
)

var (
	url, file *string
)

func init() {
	url = flag.String("m", "", "Marathon URL")
	file = flag.String("f", "", "Job file")
}

func eventBus(in <-chan RawEvent, out chan<- Event) {

	defer close(out)

	for {

		select {

		case in, ok := <-in:
			if !ok {
				break
			}

			e := new(Event)
			err := e.Unmarshal(in)

			switch {

			case err != nil && err.Error() == "Unhandled event":
				continue

			case err != nil:
				log.Println("Error parsing event:", err, in.Data)
				continue

			default:
				out <- *e

			}
		}
	}

	return
}

func main() {
	flag.Parse()
	if *url == "" {
		log.Fatal("Marathon URL (-m) is required")
	}

	raw := make(chan RawEvent, 64)
	parsed := make(chan Event, 64)
	err := EventListener(*url, raw)
	if err != nil {
		log.Fatal(err)
	}

	go eventBus(raw, parsed)

	// Do POST request here

}
