package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/namsral/flag"
)

const (
	xGoHttpInstance = "X-Go-Http-Instance"
)

var (
	address       = flag.String("address", ":8080", "Address to listen on")
	shutdownDelay = flag.Duration("shutdown-delay", time.Second, "Delay shutdown process by this duration")
)

func main() {
	flag.Parse()

	hostname, _ := os.Hostname()

	lis, err := net.Listen("tcp", *address)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(xGoHttpInstance, hostname)
		w.WriteHeader(http.StatusOK)

		_, _ = w.Write([]byte(fmt.Sprintf("Requested URI: %+s", r.URL.String())))
	})

	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if _, err := os.Stat("/tmp/not-ready"); err == nil {
			w.WriteHeader(500)
		} else {
			w.WriteHeader(200)
		}
	})

	mux.HandleFunc("/livez", func(w http.ResponseWriter, r *http.Request) {
		if _, err := os.Stat("/tmp/not-alive"); err == nil {
			w.WriteHeader(500)
		} else {
			w.WriteHeader(200)
		}
	})

	mux.HandleFunc("/sleep", func(w http.ResponseWriter, r *http.Request) {
		var sleepSeconds float64

		fmt.Sscanf(r.URL.Query().Get("sleep"), "%f", &sleepSeconds)
		if sleepSeconds > 60 {
			sleepSeconds = 60
		}
		if sleepSeconds > 0 {
			time.Sleep(time.Duration(sleepSeconds * float64(time.Second)))
		}

		w.WriteHeader(200)
		w.Write([]byte(fmt.Sprintf("%s: slept %f seconds\n", hostname, sleepSeconds)))
	})

	mux.HandleFunc("/ip", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(r.RemoteAddr))
	})

	mux.HandleFunc("/request", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-type", "text/plain")
		w.WriteHeader(200)

		w.Write([]byte(fmt.Sprintf("%s %s\r\n", r.Method, r.RequestURI)))
		r.Header.Write(w)
	})

	http := &http.Server{
		Handler: mux,
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		http.Serve(lis)
	}()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	<-sigchan

	if *shutdownDelay > 0 {
		log.Print("Delaying shutdown...")
		time.Sleep(*shutdownDelay)
	}

	log.Print("Shutting down...")

	http.Shutdown(context.Background())

	wg.Wait()
}
