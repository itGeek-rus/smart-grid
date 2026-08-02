package rest

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/itGeek-rus/smart-grid.git/internal/domain"
	"github.com/itGeek-rus/smart-grid.git/internal/usecase"
)

type APIHandler struct {
	uc *usecase.APIUseCase
}

func NewAPIHandler(uc *usecase.APIUseCase) *APIHandler {
	return &APIHandler{uc: uc}
}

func (h *APIHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/devices", h.listDevices)
	mux.HandleFunc("GET /api/v1/devices/{id}", h.getDevice)
	mux.HandleFunc("GET /api/v1/devices/{id}/telemetry", h.getTelemetry)
	mux.HandleFunc("GET /api/v1/devices/{id}/telemetry/latest", h.getLatest)
	mux.HandleFunc("GET /api/v1/devices/{id}/alerts", h.listAlerts)
	mux.HandleFunc("POST /api/v1/devices/{id}/commands", h.sendCommand)
}

func (h *APIHandler) listDevices(w http.ResponseWriter, r *http.Request) {
	zone := r.URL.Query().Get("zone")
	var (
		items []domain.Device
		err   error
	)
	if zone == "" {
		items, err = h.uc.ListAllDevices(r.Context())
	} else {
		items, err = h.uc.ListDevices(r.Context(), zone)
	}
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *APIHandler) getDevice(w http.ResponseWriter, r *http.Request) {
	d, err := h.uc.GetDevice(r.Context(), r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, d)
}

func (h *APIHandler) getTelemetry(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	from, _ := time.Parse(time.RFC3339, q.Get("from"))
	to, _ := time.Parse(time.RFC3339, q.Get("to"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	items, err := h.uc.GetTelemetry(r.Context(), r.PathValue("id"), from, to, limit)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *APIHandler) getLatest(w http.ResponseWriter, r *http.Request) {
	t, err := h.uc.GetLatestTelemetry(r.Context(), r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (h *APIHandler) listAlerts(w http.ResponseWriter, r *http.Request) {
	items, err := h.uc.ListOpenAlerts(r.Context(), r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

type commandRequest struct {
	Command string            `json:"command"`
	Params  map[string]string `json:"params"`
}

func (h *APIHandler) sendCommand(w http.ResponseWriter, r *http.Request) {
	var req commandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Command == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	evt := domain.DeviceCommandEvent{
		Envelope: domain.Envelope{
			EventID:       uuid.NewString(),
			EventType:     domain.EventTypeDeviceCommand,
			SchemaVersion: 1,
			OccurredAt:    time.Now().UTC(),
		},
		CommandID: uuid.NewString(),
		DeviceID:  r.PathValue("id"),
		Command:   req.Command,
		Params:    req.Params,
		IssuedAt:  time.Now().UTC(),
	}
	if err := h.uc.SendCommand(r.Context(), evt); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, evt)
}

func writeErr(w http.ResponseWriter, err error) {
	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
}
