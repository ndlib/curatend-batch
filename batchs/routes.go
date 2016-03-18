package batchs

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	_ "net/http/pprof" // for pprof server

	"github.com/julienschmidt/httprouter"
)

// RESTServer holds the configuration for a Batch REST API server.
//
// Set all the public fields and then call Run. Run will listen on the given
// port and handle requests. At the moment there is no maximum simultaneous
// request limit. Do not change any fields after calling Run.
//
type RESTServer struct {
	// Port number to run bendo on. defaults to 15000
	PortNumber string
	QueuePath string
}

// the number of active commits onto tape we allow at a given time
const MaxConcurrentCommits = 2

// Run initializes and starts all the goroutines used by the server. It then
// blocks listening for and handling http requests.
func (s *RESTServer) Run() error {
	log.Println("==========")
	log.Printf("Starting HTTP Server")
	log.Printf("PortNumber = %s", s.PortNumber)
	log.Printf("QueuePath = %s", s.QueuePath)

	err := http.ListenAndServe(":" + s.PortNumber, s.addRoutes()) 

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (s *RESTServer) addRoutes() http.Handler {
	var routes = []struct {
		method  string
		route   string
		handler httprouter.Handle
	}{
		// the /blob/* routes can be removed. they are functionally the
		// same as /item/@blob/*
		{"GET", "/blob/:id/:bid", NotImplementedHandler},
		{"HEAD", "/blob/:id/:bid", NotImplementedHandler},
		{"GET", "/item/:id/*slot", NotImplementedHandler},
		{"HEAD", "/item/:id/*slot", NotImplementedHandler},
		{"GET", "/item/:id", NotImplementedHandler},

		// all the transaction things.
		{"POST", "/item/:id/transaction", NotImplementedHandler},
		{"GET", "/transaction", NotImplementedHandler},
		{"GET", "/transaction/:tid", NotImplementedHandler},
		{"POST", "/transaction/:tid/cancel", NotImplementedHandler}, //keep?
	}

	r := httprouter.New()
	for _, route := range routes {
		r.Handle(route.method,
			route.route,
			logWrapper(route.handler))
	}
	return r
}

// General route handlers and convinence functions

// NotImplementedHandler will return a 501 not implemented error.
func NotImplementedHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.WriteHeader(http.StatusNotImplemented)
	fmt.Fprintf(w, "Not Implemented\n")
}

// writeHTMLorJSON will either return val as JSON or as rendered using the
// given template, depending on the request header "Accept-Encoding".
func writeHTMLorJSON(w http.ResponseWriter,
	r *http.Request,
	tmpl *template.Template,
	val interface{}) {

	if r.Header.Get("Accept-Encoding") == "application/json" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(val)
		return
	}
	tmpl.Execute(w, val)
}

// logWrapper takes a handler and returns a handler which does the same thing,
// after first logging the request URL.
func logWrapper(handler httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		log.Println(r.Method, r.URL)
		handler(w, r, ps)
	}
}
