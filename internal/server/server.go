package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/mihir/transactions/internal/handler"
)

func Start(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.HandleRoot)

	s := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Println("listening on", addr)
	return s.ListenAndServe()
}
