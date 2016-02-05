package main

import (
	"log"

	"github.com/kylebrandt/annotate/web"
	"github.com/kylebrandt/annotate/backend"
)



func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	b, err := backend.NewElastic([]string{"http://ny-devlogstash04:9200"}, "annotate")
	if err != nil {
		log.Fatal(err)
	}
	backends := []backend.Backend{b}
	for _, b := range backends {
		if err := b.InitBackend(); err != nil {
			log.Fatal(err)
		}
	}
	go func() { log.Fatal(web.Listen(":8080", backends)) }()
	select {}
}