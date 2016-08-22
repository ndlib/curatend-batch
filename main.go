package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ndlib/curatend-batch/batchs"
)

type Reopener interface {
	Reopen()
}

type loginfo struct {
	name string
	f    *os.File
}

func NewReopener(filename string) *loginfo {
	return &loginfo{name: filename}
}

func (li *loginfo) Reopen() {
	if li.name == "" {
		return
	}
	if li.f != nil {
		log.Println("Reopening Log files")
	}
	newf, err := os.OpenFile(li.name, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(newf)
	if li.f != nil {
		li.f.Close()
	}
	li.f = newf
}

func signalHandler(sig <-chan os.Signal, logw Reopener) {
	for s := range sig {
		if s == syscall.SIGCHLD {
			// each call to exec will generate a SIGCHLD when the child process exits.
			// We don't care about them.
			continue
		}
		log.Println("Received signal", s)
		switch s {
		case syscall.SIGUSR1:
			logw.Reopen()
		case syscall.SIGINT, syscall.SIGTERM:
			log.Println("Exiting")
			if pidfilename != "" {
				// we don't care if there is an error
				os.Remove(pidfilename)
			}
			os.Exit(1)
		}
	}
}

func writePID(fname string) {
	f, err := os.Create(fname)
	if err != nil {
		log.Printf("Error writing PID to file '%s': %s\n", fname, err.Error())
		return
	}
	pid := os.Getpid()
	fmt.Fprintf(f, "%d", pid)
	f.Close()
}

var (
	pidfilename string
)

func main() {
	var (
		logw        Reopener
		logfilename = flag.String("log", "", "name of log file")
		portNumber  = flag.String("port", "15000", "Port Number of httpd service")
		showVersion = flag.Bool("version", false, "Display binary version")
		queuepath   = flag.String("queue", "test", "path to the queue directory")
		taskpath    = flag.String("tasks", "tasks", "path to the task commands")
	)

	flag.StringVar(&pidfilename, "pid", "", "file to store pid of server")
	flag.Parse()

	if *showVersion {
		fmt.Printf("batchs version %s\n", version)
		return
	}

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	logw = NewReopener(*logfilename)
	logw.Reopen()
	log.Println("-----Starting Server")

	log.Println("queuepath =", *queuepath)
	log.Println("taskpath =", *taskpath)
	log.Println("PortNumber = ", *portNumber)

	if pidfilename != "" {
		writePID(pidfilename)
	}

	sig := make(chan os.Signal, 5)
	signal.Notify(sig)
	go signalHandler(sig, logw)

	fs := batchs.NewFileQueue(*queuepath)
	ctx := batchs.NewContext(fs, *taskpath, version)

	hs := batchs.RESTServer{
		QueuePath:  fs,
		PortNumber: *portNumber,
		Version: version,
	}

	//Start the HTTPD Server thread

	go hs.Run()

	// Start the batchs server thread

	err := ctx.Run()

	if pidfilename != "" {
		// we don't care if there is an error
		os.Remove(pidfilename)
	}

	if err != nil {
		os.Exit(2)
	}
}
