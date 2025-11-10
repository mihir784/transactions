package service

import (
	"encoding/json"
	"net/http"
)

type Service struct {
	dbPool *database.dbPool
}

type statusResp struct {
	Status string `json:"status"`
}

func HandleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(statusResp{Status: "ok"})
}
