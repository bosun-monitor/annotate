package backend

import (
	"github.com/kylebrandt/annotate"
	elastic "gopkg.in/olivere/elastic.v3"
)

type Backend interface {
	InsertAnnotation(a *annotate.Annotation) error
	InitBackend(table string) error
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
	_, err := e.Index().Index(e.index).Do()
	return err
}

func (e *Elastic) InitBackend(table string) error {
	exists, err := e.IndexExists(table).Do()
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	res, err := e.CreateIndex(table).Do()
	if res.Acknowledged && err != nil {
		return nil
	}
	// TODO Create a Elastic Mapping
	return err
}
