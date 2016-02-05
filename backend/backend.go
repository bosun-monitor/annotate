package backend

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/kylebrandt/annotate"
	elastic "gopkg.in/olivere/elastic.v3"
)

type Backend interface {
	InsertAnnotation(a *annotate.Annotation) error
	GetAnnotation(id string) (*annotate.Annotation, error)
	GetAnnotations(start, end time.Time) (annotate.Annotations, error)
	InitBackend() error
}

const docType = "annotation"

type Elastic struct {
	*elastic.Client
	index string
}

func NewElastic(urls []string, index string) (*Elastic, error) {
	e, err := elastic.NewClient(elastic.SetURL(urls...))
	return &Elastic{e, index}, err
}

func (e *Elastic) InsertAnnotation(a *annotate.Annotation) error {
	_, err := e.Index().Index(e.index).BodyJson(a).Id(a.Id).Type(docType).Do()
	return err
}

func (e *Elastic) GetAnnotation(id string) (*annotate.Annotation, error) {
	a := annotate.Annotation{}
	res, err := e.Get().Index(e.index).Type(docType).Id(id).Do()
	if err != nil {
		return &a, fmt.Errorf("%v: %v", err, res.Error.Reason)
	}
	if err := json.Unmarshal(*res.Source, &a); err != nil {
		return &a, err
	}
	return &a, nil
}

func (e *Elastic) GetAnnotations(start, end time.Time) (annotate.Annotations, error) {
	annotations := annotate.Annotations{}
	startQ := elastic.NewRangeQuery(annotate.StartDate).Gte(start)
	endQ := elastic.NewRangeQuery(annotate.EndDate).Lte(end)
	and := elastic.NewBoolQuery().Must(startQ, endQ)
	res, err := e.Search(e.index).Query(and).Do()
	if err != nil {
		return annotations, fmt.Errorf("%v: %v", err, res.Error.Reason)
	}
	var aType annotate.Annotation
	for _, item := range res.Each(reflect.TypeOf(aType)) {
		a := item.(annotate.Annotation)
		annotations = append(annotations, a)
	}
	return annotations, nil
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
