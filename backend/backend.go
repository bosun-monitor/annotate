package backend

import (
	"github.com/kylebrandt/annotate"
	elastic "gopkg.in/olivere/elastic.v3"
)

type Backend interface {
	InsertAnnotation(a *annotate.Annotation) error
	InitBackend() error
}

type Elastic struct {
	*elastic.Client
	index string
}

func NewElastic(urls []string, index string) (*Elastic, error) {
	e, err := elastic.NewClient(elastic.SetURL(urls...))
	return &Elastic{e, index}, err
}

func (e *Elastic) InsertAnnotation(a *annotate.Annotation) error {
	_, err := e.Index().Index(e.index).BodyJson(a).Type("annotation").Do()
	return err
}

func (e *Elastic) InitBackend() error {
	exists, err := e.IndexExists(e.index).Do()
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	res, err := e.CreateIndex(e.index).Do()
	if res.Acknowledged && err != nil {
		return nil
	}
	// TODO Create a Elastic Mapping
	return err
}
