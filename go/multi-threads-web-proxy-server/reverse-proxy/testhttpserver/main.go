package main

import (
	"log"
	"net/http"
	"net/http/httptest"
)

var hosts = []string{"localhost:9091", "localhost:9092"}

func main() {
	for _, host := range hosts {
		host := host
		go serve(host)
	}

	select {}
}

func serve(addr string) {

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("你好，这里是 " + addr)); err != nil {
			log.Fatalln(err)
		}
	})

	mux := http.NewServeMux()
	mux.Handle("/hello", handler)
	err := http.ListenAndServe(addr, mux)
	if err != nil {
		log.Fatalln(err)
	}
	httptest.NewServer(handler)
}
