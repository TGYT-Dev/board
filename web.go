package main

import (
	"net/http"
)

func serve() {
	http.Handle("/", http.FileServer(http.Dir("./web")))
	http.ListenAndServe(":3001", nil)
}
