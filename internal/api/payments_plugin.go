package api

import (
	"context"
	"net/http"

	"xlpanel/internal/plugins"
)

type pluginChargeRequest struct {
	Provider   string            `json:"provider"`
	TenantID   string            `json:"tenant_id"`
	CustomerID string            `json:"customer_id"`
	InvoiceID  string            `json:"invoice_id"`
	Amount     float64           `json:"amount"`
	Currency   string            `json:"currency"`
	Metadata   map[string]string `json:"metadata"`
}

type pluginChargeResponse struct {
	Provider      string `json:"provider"`
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
	Message       string `json:"message"`
}

func (s *Server) handlePluginCharge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req pluginChargeRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Amount <= 0 {
		writeError(w, http.StatusBadRequest, "amount must be positive")
		return
	}
	if req.Currency == "" {
		req.Currency = "USD"
	}
	response, err := s.paymentsService.Charge(context.Background(), req.Provider, plugins.PaymentRequest{
		TenantID:   req.TenantID,
		CustomerID: req.CustomerID,
		InvoiceID:  req.InvoiceID,
		Amount:     req.Amount,
		Currency:   req.Currency,
		Metadata:   req.Metadata,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, pluginChargeResponse{
		Provider:      req.Provider,
		TransactionID: response.TransactionID,
		Status:        response.Status,
		Message:       response.Message,
	})
}
