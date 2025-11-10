package handler

import (
	"fmt"
	"net/http"
	"time"
)

type Handler struct {
	dbPool *database.dbPool
}


func (s *Handler) Start(addr string, dbPool *database.dbPool) error {
	service := &service.Service{dbPool: dbPool}
	mux := http.NewServeMux()
	mux.HandleFunc("/", service.HandleRoot)
	mux.HandleFunc("/accounts", service.HandleAccounts)
	mux.HandleFunc("/accounts/", service.HandleAccountByID)
	mux.HandleFunc("/transactions", service.HandleTransactions)

	s := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Println("listening on", addr)
	return s.ListenAndServe()
}
