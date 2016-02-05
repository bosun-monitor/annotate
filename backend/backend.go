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
	GetAnnotations(start, end *time.Time, source, host, creationUser, owner, category string) (annotate.Annotations, error)
	GetFieldValues(field string) ([]string, error)
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
	if id != "" {
		return &a, fmt.Errorf("must provide id")
	}
	res, err := e.Get().Index(e.index).Type(docType).Id(id).Do()
	if err != nil {
		return &a, fmt.Errorf("%v: %v", err, res.Error.Reason)
	}
	if err := json.Unmarshal(*res.Source, &a); err != nil {
		return &a, err
	}
	return &a, nil
}

func (e *Elastic) GetAnnotations(start, end *time.Time, source, host, creationUser, owner, category string) (annotate.Annotations, error) {
	annotations := annotate.Annotations{}
	s := elastic.NewSearchSource()
	if start != nil && end != nil {
		startQ := elastic.NewRangeQuery(annotate.StartDate).Gte(start)
		endQ := elastic.NewRangeQuery(annotate.EndDate).Lte(end)
		s = s.Query(elastic.NewBoolQuery().Must(startQ, endQ))
	}
	if source != "" {
		s = s.Query(elastic.NewTermQuery(annotate.Source, source))
	}
	if host != "" {
		s = s.Query(elastic.NewTermQuery(annotate.Host, host))
	}
	if creationUser != "" {
		s = s.Query(elastic.NewTermQuery(annotate.CreationUser, creationUser))
	}
	if owner != "" {
		s = s.Query(elastic.NewTermQuery(annotate.Owner, owner))
	}
	if category != "" {
		s = s.Query(elastic.NewTermQuery(annotate.Category, category))
	}
	res, err := e.Search(e.index).Query(s).Do()
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

func (e *Elastic) GetFieldValues(field string) ([]string, error) {
	terms := []string{}
	switch field {
		case annotate.Source, annotate.Host, annotate.CreationUser, annotate.Owner, annotate.Category:
			//continue
		default:
			return terms, fmt.Errorf("invalid field %v", field)
	}
	termsAgg := elastic.NewTermsAggregation().Field(field)
	res, err := e.Search(e.index).Aggregation(field, termsAgg).Do()
	if err != nil {
		return terms, err
	}
	b, found := res.Aggregations.Terms(field)
	if !found {
		return terms, fmt.Errorf("expected aggregation %v not found in result", field)
	}
	for _, bucket := range b.Buckets {
		if v, ok := bucket.Key.(string); ok {
			terms = append(terms, v)
		}
	}
	return terms, nil
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
