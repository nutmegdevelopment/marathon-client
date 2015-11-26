package main

import (
	"errors"
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

func eventHandler(ch <-chan RawEvent) (result string, err error) {

	for {

		select {

		case in, ok := <-ch:
			if !ok {
				err = errors.New("Error: event channel closed")
				break
			}
			log.Println("RAW INPUT:")
			log.Println(in.Name)
			log.Println(string(in.Data))
			e := new(Event)
			err = e.Unmarshal(in)

			switch {

			case err != nil && err.Error() == "Unhandled event":
				log.Println(err)
				continue

			case err != nil:
				break

			default:
				log.Println("\n\n")
				log.Println(e)
				log.Println("\n\n")

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

	events := make(chan RawEvent)
	err := EventListener(*url, events)
	if err != nil {
		log.Fatal(err)
	}

	// Do POST request here

	res, err := eventHandler(events)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(res)

}
