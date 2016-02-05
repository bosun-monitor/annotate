package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/kylebrandt/annotate"
	"github.com/kylebrandt/annotate/backend"
)

func Listen(listenAddr string, b []backend.Backend) error {
	backends = b
	router.HandleFunc("/", InsertAnnotation).Methods("POST")
	router.HandleFunc("/query", GetAnnotations).Methods("GET")
	router.HandleFunc("/{id}", GetAnnotation).Methods("GET")
	http.Handle("/", router)
	return http.ListenAndServe(listenAddr, nil)
}

//Web Section
var (
	router   = mux.NewRouter()
	backends = []backend.Backend{}
)

func InsertAnnotation(w http.ResponseWriter, req *http.Request) {
	var a annotate.Annotation
	d := json.NewDecoder(req.Body)
	err := d.Decode(&a)
	if err != nil {
		serveError(w, err)
		return
	}
	if a.IsOneTimeSet() {
		a.MatchTimes()
	}
	if a.IsTimeNotSet() {
		a.SetNow()
	}
	err = a.ValidateTime()
	if err != nil {
		serveError(w, err)
	}
	if a.Id == "" { //if Id isn't set, this is a new Annotation
		a.SetGUID()
	}
	for _, b := range backends {
		err := b.InsertAnnotation(&a)
		//TODO Collect errors and insert into the backends that we can
		if err != nil {
			serveError(w, err)
		}
	}
	log.Println(a)
	err = json.NewEncoder(w).Encode(a)
	if err != nil {
		serveError(w, err)
	}
	w.Header().Set("Content-Type", "application/json")
	return
}

func GetAnnotation(w http.ResponseWriter, req *http.Request) {
	var a *annotate.Annotation
	var err error
	id := mux.Vars(req)["id"]
	for _, b := range backends {
		a, err = b.GetAnnotation(id)
		//TODO Collect errors and insert into the backends that we can
		if err != nil {
			serveError(w, err)
		}
	}
	err = json.NewEncoder(w).Encode(a)
	if err != nil {
		serveError(w, err)
	}
	w.Header().Set("Content-Type", "application/json")
	return
}

func GetAnnotations(w http.ResponseWriter, req *http.Request) {
	var a annotate.Annotations
	var startT *time.Time
	var endT *time.Time
	var err error
	w.Header().Set("Content-Type", "application/json")

	// Time
	start := req.URL.Query().Get(annotate.StartDate)
	end := req.URL.Query().Get(annotate.EndDate)
	if start != "" {
		s, err := time.Parse(time.RFC3339Nano, start)
		if err != nil {
			serveError(w, fmt.Errorf("error parsing StartDate %v: %v", start, err))
		}
		startT = &s
	}
	if end != "" {
		e, err := time.Parse(time.RFC3339Nano, end)
		if err != nil {
			serveError(w, fmt.Errorf("error parsing EndDate %v: %v", end, err))
		}
		endT = &e
	}

	// Other Fields
	source := req.URL.Query().Get(annotate.Source)
	host := req.URL.Query().Get(annotate.Host)
	creationUser := req.URL.Query().Get(annotate.CreationUser)
	owner := req.URL.Query().Get(annotate.Owner)

	// Execute
	for _, b := range backends {
		a, err = b.GetAnnotations(startT, endT, source, host, creationUser, owner)
		//TODO Collect errors and insert into the backends that we can
		if err != nil {
			serveError(w, err)
		}
	}

	// Encode
	err = json.NewEncoder(w).Encode(a)
	if err != nil {
		serveError(w, err)
	}
	return
}

func serveError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
