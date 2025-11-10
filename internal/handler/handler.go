package handler

import (
	"encoding/json"
	"net/http"
)

type statusResp struct {
	Status string `json:"status"`
}

func HandleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(statusResp{Status: "ok"})
}
