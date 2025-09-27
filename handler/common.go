package handler

import (
	"encoding/json"
	"net/http"
)

// respondJSON makes the response with payload as json format
func sendResponse(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(response))
}

type APIResponse interface {
	Dispatch()
}

type SuccessResponse struct {
	RW      http.ResponseWriter
	Status  int
	Data    any
	Message string
}

type ErrorResponse struct {
	RW     http.ResponseWriter
	Status int
	Error  any
}

type SuccessJSON struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type ErrorJSON struct {
	Status int `json:"status"`
	Error  any `json:"error"`
}

func (r *SuccessResponse) Dispatch() {
	sendResponse(r.RW, r.Status, SuccessJSON{
		Status:  r.Status,
		Message: r.Message,
		Data:    r.Data,
	})
}

func (r *ErrorResponse) Dispatch() {
	sendResponse(r.RW, r.Status, ErrorJSON{
		Status: r.Status,
		Error:  r.Error,
	})
}
