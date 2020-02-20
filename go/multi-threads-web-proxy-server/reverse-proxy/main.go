package main

import (
	"log"
	"net/http"
	"net/url"
)

func main() {


	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		proxy := NewHTTPReverseProxy([]*url.URL{
			{
				Scheme:"http",
				Host:"localhost:9001",
			},
			{
				Scheme:"http",
				Host:"localhost:9002",
			},
		})
		proxy.ServeHTTP(w, r)
	})

	if err := http.ListenAndServe(":9000", nil); err != nil {
		log.Fatalln(err)
	}
}
