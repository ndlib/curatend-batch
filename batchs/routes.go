package batchs

import (
	"fmt"
	"log"
	"net/http"
	"time"

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
	QueuePath  *FileQueue
	Version    string
}

// Run initializes and starts all the goroutines used by the server. It then
// blocks listening for and handling http requests. This function never returns.
func (server *RESTServer) Run() {
	if server.PortNumber == "" {
		server.PortNumber = "15000"
	}

	for {
		err := http.ListenAndServe(":"+server.PortNumber, server.addRoutes())

		if err != nil {
			log.Println(err)
			time.Sleep(5 * time.Second) // duration is arbitrary
		}
	}
}

// The HTTP interface
func (server *RESTServer) addRoutes() http.Handler {
	var routes = []struct {
		method  string
		route   string
		handler httprouter.Handle
	}{
		{"GET", "/", server.WelcomeHandler},
		{"HEAD", "/", server.WelcomeHandler},
		{"GET", "/jobs", server.GetJobsHandler},
		{"GET", "/jobs/:id", server.GetJobIDHandler},
		{"HEAD", "/jobs/:id", server.GetJobIDHandler},
		{"GET", "/jobs/:id/files/*path", server.GetJobIDFileHandler},
		{"HEAD", "/jobs/:id/files/*path", server.GetJobIDFileHandler},
		{"PUT", "/jobs/:id", server.PutJobIDHandler},
		{"DELETE", "/jobs/:id", server.DeleteJobIDHandler},
		{"POST", "/jobs/:id/queue", server.SubmitJobIDHandler},
		{"PUT", "/jobs/:id/files/*path", server.PutJobIDFileHandler},
		{"DELETE", "/jobs/:id/files/*path", server.DeleteJobIDFileHandler},
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
