package v1

import (
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	authmw "github.com/negeek/solar-sphere/solar-sentinel/middlewares/v1"
	service "github.com/negeek/solar-sphere/solar-sentinel/service/v1"
	"github.com/negeek/solar-sphere/solar-spectrum/httpapi"
)

type Handler struct {
	devices    *service.DeviceService
	irradiance *service.IrradianceService
}

func NewHandler(devices *service.DeviceService, irradiance *service.IrradianceService) *Handler {
	return &Handler{devices: devices, irradiance: irradiance}
}

type createDeviceRequest struct {
	Name string `json:"name"`
}

// CreateDevice registers a new device owned by the authenticated caller. A
// user can call this as many times as they like to register more devices —
// there is no admin gate.
func (h *Handler) CreateDevice(w http.ResponseWriter, r *http.Request) {
	var req createDeviceRequest
	if err := httpapi.Unmarshall(r.Body, &req); err != nil {
		httpapi.JsonResponse(w, false, http.StatusBadRequest, err.Error(), nil)
		return
	}

	device, err := h.devices.CreateDevice(r.Context(), service.CreateDeviceInput{
		Name:  req.Name,
		Owner: authmw.EmailFromContext(r.Context()),
	})
	if err != nil {
		httpapi.JsonResponse(w, false, http.StatusBadRequest, err.Error(), nil)
		return
	}

	httpapi.JsonResponse(w, true, http.StatusCreated, "Successfully created device", map[string]interface{}{"device_id": device.ID})
}

// DownloadSolarIrrData streams a device's readings as CSV, after confirming
// the authenticated caller owns the device.
func (h *Handler) DownloadSolarIrrData(w http.ResponseWriter, r *http.Request) {
	deviceID := mux.Vars(r)["device_id"]
	requester := authmw.EmailFromContext(r.Context())

	rows, err := h.irradiance.Export(r.Context(), deviceID, requester)
	if err != nil {
		httpapi.JsonResponse(w, false, statusForExportErr(err), err.Error(), nil)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.csv", deviceID))

	writer := csv.NewWriter(w)
	defer writer.Flush()
	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return
		}
	}
}

func statusForExportErr(err error) int {
	switch {
	case errors.Is(err, service.ErrDeviceNotFound):
		return http.StatusNotFound
	case errors.Is(err, service.ErrNotDeviceOwner):
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}
