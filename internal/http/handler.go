package http

import (
	"encoding/json"
	"html/template"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"myhomeapp/internal/api"
)

type Handler struct {
	asekoClient *api.AsekoClient
	templates   *template.Template
}

type MeasurementData struct {
	Value     float64
	Unit      string
	UpdatedAt time.Time
}

type MeasurementsPageData struct {
	Redox            MeasurementData
	PH               MeasurementData
	WaterTemperature MeasurementData
	WaterFlow        MeasurementData
}

func NewHandler(asekoClient *api.AsekoClient) *Handler {
	templates := template.Must(template.ParseGlob("web/templates/*.html"))
	return &Handler{
		asekoClient: asekoClient,
		templates:   templates,
	}
}

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	unit := h.asekoClient.GetSelectedUnit()
	if unit == nil {
		http.Error(w, "No unit selected", http.StatusInternalServerError)
		return
	}

	// Extract measurements
	var data MeasurementsPageData
	for _, measurement := range unit.Data.UnitBySerialNumber.Measurements {
		switch measurement.Type {
		case "REDOX":
			data.Redox = MeasurementData{
				Value:     measurement.Value,
				Unit:      measurement.Unit,
				UpdatedAt: measurement.UpdatedAt,
			}
		case "PH":
			data.PH = MeasurementData{
				Value:     measurement.Value,
				Unit:      measurement.Unit,
				UpdatedAt: measurement.UpdatedAt,
			}
		case "WATER_TEMPERATURE":
			data.WaterTemperature = MeasurementData{
				Value:     measurement.Value,
				Unit:      measurement.Unit,
				UpdatedAt: measurement.UpdatedAt,
			}
		case "WATER_FLOW_TO_PROBES":
			data.WaterFlow = MeasurementData{
				Value:     measurement.Value,
				Unit:      measurement.Unit,
				UpdatedAt: measurement.UpdatedAt,
			}
		}
	}

	if err := h.templates.ExecuteTemplate(w, "measurements.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) GetStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

func (h *Handler) GetUnits(w http.ResponseWriter, r *http.Request) {
	units := h.asekoClient.GetUnitList()
	if units == nil {
		http.Error(w, "No units available", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(units)
}

func (h *Handler) GetUnitDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serialNumber := vars["serialNumber"]
	
	if serialNumber == "" {
		http.Error(w, "Serial number required", http.StatusBadRequest)
		return
	}
	
	// Select the unit first
	if err := h.asekoClient.SelectUnit(serialNumber); err != nil {
		http.Error(w, "Failed to select unit: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	unit := h.asekoClient.GetSelectedUnit()
	if unit == nil {
		http.Error(w, "Unit not found", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(unit)
}
