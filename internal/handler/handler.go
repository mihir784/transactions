package handler

import (
	"fmt"
	"net/http"
	"time"
	
	"github.com/jackc/pgx/v5/pgxpool"
	"transactions/internal/service"
)

type Handler struct {
	dbPool *pgxpool.Pool
}


func (h *Handler) Start(addr string, dbPool *pgxpool.Pool) error {
	service := &service.Service{DbPool: dbPool}
	mux := http.NewServeMux()
	mux.HandleFunc("/", service.HandleRoot)
	mux.HandleFunc("/accounts", service.HandleAccounts)
	mux.HandleFunc("/accounts/", service.HandleAccountByID)
	mux.HandleFunc("/transactions", service.HandleTransactions)

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Println("listening on", addr)
	return server.ListenAndServe()
}
