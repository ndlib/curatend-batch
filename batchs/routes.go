package batchs

import (
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// RESTServer holds the configuration for a Batch REST API server.
//
// Set all the public fields and then call Run. Run will listen on the given
// port and handle requests. At the moment there is no maximum simultaneous
// request limit. Do not change any fields after calling Run.
type RESTServer struct {
	// Port number to run bendo on. defaults to 15000
	PortNumber string
	QueuePath  *fileQueue
	Version string
}

// Run initializes and starts all the goroutines used by the server. It then
// blocks listening for and handling http requests.
func (s *RESTServer) Run() error {
	if s.PortNumber == "" {
		s.PortNumber = "15000"
	}

	err := http.ListenAndServe(":"+s.PortNumber, s.addRoutes())

	if err != nil {
		log.Println(err)
	}

	return err
}

func (server *RESTServer) addRoutes() http.Handler {
	var routes = []struct {
		method  string
		route   string
		handler httprouter.Handle
	}{
		{"GET", "/", server.WelcomeHandler},
		{"HEAD", "/", server.WelcomeHandler},
		{"GET", "/jobs", server.GetJobsHandler},
		{"GET", "/jobs/:id", server.GetJobIdHandler},
		{"HEAD", "/jobs/:id", server.GetJobIdHandler},
		{"GET", "/jobs/:id/files/*path", server.GetJobIdFileHandler},
		{"HEAD", "/jobs/:id/files/*path", server.GetJobIdFileHandler},
		{"PUT", "/jobs/:id", server.PutJobIdHandler},
		{"DELETE", "/jobs/:id", server.DeleteJobIdHandler},
		{"POST", "/jobs/:id/queue", server.SubmitJobIdHandler},
		{"PUT", "/jobs/:id/files/*path", server.PutJobIdFileHandler},
		{"DELETE", "/jobs/:id/files/*path", server.DeleteJobIdFileHandler},
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

// logWrapper takes a handler and returns a handler which does the same thing,
// after first logging the request URL.
func logWrapper(handler httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		log.Println(r.Method, r.URL)
		handler(w, r, ps)
	}
}
