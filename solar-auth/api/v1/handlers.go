package v1

import (
	"net/http"
	"time"

	service "github.com/negeek/solar-sphere/solar-auth/service/v1"
	"github.com/negeek/solar-sphere/solar-spectrum/httpapi"
)

type Handler struct {
	auth *service.AuthService
}

func NewHandler(auth *service.AuthService) *Handler {
	return &Handler{auth: auth}
}

type joinRequest struct {
	Email string `json:"email"`
	// ExpiresInDays is optional; omitted or zero means the issued key never
	// expires on its own.
	ExpiresInDays int `json:"expires_in_days,omitempty"`
}

type rotateRequest struct {
	Key           string `json:"key"`
	Email         string `json:"email"`
	ExpiresInDays int    `json:"expires_in_days,omitempty"`
}

// Join signs a new user up and issues their first access key.
func (h *Handler) Join(w http.ResponseWriter, r *http.Request) {
	var req joinRequest
	if err := httpapi.Unmarshall(r.Body, &req); err != nil {
		httpapi.JsonResponse(w, false, http.StatusBadRequest, err.Error(), nil)
		return
	}

	result, err := h.auth.SignUp(r.Context(), service.SignUpInput{
		Email:     req.Email,
		ExpiresIn: daysToDuration(req.ExpiresInDays),
	})
	if err != nil {
		httpapi.JsonResponse(w, false, statusFor(err), err.Error(), nil)
		return
	}

	httpapi.JsonResponse(w, true, http.StatusCreated, "Successfully joined", map[string]interface{}{
		"email":      result.Email,
		"access_key": result.AccessKey,
	})
}

// NewAccessKey rotates a user's access key: their previous key must be
// presented and is revoked as part of issuing the new one.
func (h *Handler) NewAccessKey(w http.ResponseWriter, r *http.Request) {
	var req rotateRequest
	if err := httpapi.Unmarshall(r.Body, &req); err != nil {
		httpapi.JsonResponse(w, false, http.StatusBadRequest, err.Error(), nil)
		return
	}

	result, err := h.auth.RotateKey(r.Context(), service.RotateKeyInput{
		OldKey:    req.Key,
		Email:     req.Email,
		ExpiresIn: daysToDuration(req.ExpiresInDays),
	})
	if err != nil {
		httpapi.JsonResponse(w, false, statusFor(err), err.Error(), nil)
		return
	}

	httpapi.JsonResponse(w, true, http.StatusCreated, "Successfully changed access key", map[string]interface{}{
		"email":      result.Email,
		"access_key": result.AccessKey,
	})
}

func daysToDuration(days int) time.Duration {
	if days <= 0 {
		return 0
	}
	return time.Duration(days) * 24 * time.Hour
}

func statusFor(err error) int {
	switch err {
	case service.ErrInvalidEmail, service.ErrInvalidAccessKey, service.ErrEmailMismatch, service.ErrAlreadyRevoked:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
