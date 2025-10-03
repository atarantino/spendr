package handlers

import (
	"encoding/json"
	"net/http"

	"spendr/internal/database"
)

type HealthHandler struct {
	db database.Service
}

func NewHealthHandler(db database.Service) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	jsonResp, _ := json.Marshal(h.db.Health())
	_, _ = w.Write(jsonResp)
}
