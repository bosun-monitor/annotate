package main

import (
	"log"

	"github.com/BurntSushi/toml"
	"github.com/kylebrandt/annotate/backend"
	"github.com/kylebrandt/annotate/web"
)

type Conf struct {
	ListenAddress   string
	ElasticClusters []ElasticCluster
}

type ElasticCluster struct {
	Servers []string // i.w. http://ny-elastic01:9200
	Index   string
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	var c Conf
	if _, err := toml.DecodeFile("./config.toml", &c); err != nil {
		log.Fatal("failed to decode config file: ", err)
	}
	backends := []backend.Backend{}
	for _, eCluster := range c.ElasticClusters {
		b, err := backend.NewElastic(eCluster.Servers, eCluster.Index)
		if err != nil {
			log.Fatal(err)
		}
		backends = append(backends, b)
	}
	for _, b := range backends {
		if err := b.InitBackend(); err != nil {
			log.Fatal(err)
		}
	}
	go func() { log.Fatal(web.Listen(":8080", backends)) }()
	select {}
}
