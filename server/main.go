package main

import (
	"log"

	"github.com/kylebrandt/annotate/web"
	"github.com/kylebrandt/annotate/backend"
)

func main() {
	b, err := backend.NewElastic([]string{"http://ny-devlogstash04:9200"}, "annotate")
	if err != nil {
		log.Fatal(err)
	}
	backends := []backend.Backend{b}
	go func() { log.Fatal(web.Listen(":8080", backends)) }()
	select {}
}

