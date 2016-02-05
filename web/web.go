package web

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kylebrandt/annotate"
	"github.com/kylebrandt/annotate/backend"

)

func Listen(listenAddr string, b []backend.Backend) error {
	backends = b
	router.HandleFunc("/", InsertAnnotation).Methods("POST")
	http.Handle("/", router)
	return http.ListenAndServe(listenAddr, nil)
}

//Web Section
var (
	router = mux.NewRouter()
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

func serveError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
