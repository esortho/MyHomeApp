package api

import (
	"html/template"
	"net/http"
)

type MeasurementData struct {
	Redox            struct{ Value string }
	PH               struct{ Value string }
	WaterTemperature struct{ Value string }
	WaterFlow        struct{ Value string }
}

func (c *AsekoClient) HandleMeasurements(w http.ResponseWriter, r *http.Request) {
	unit := c.GetSelectedUnit()
	if unit == nil {
		http.Error(w, "No unit selected", http.StatusInternalServerError)
		return
	}

	data := MeasurementData{}

	// Extract values from status values
	for _, sv := range unit.Data.UnitBySerialNumber.StatusValues.Primary {
		switch sv.Type {
		case "REDOX":
			data.Redox.Value = sv.Center.Value
		case "PH":
			data.PH.Value = sv.Center.Value
		case "WATER_TEMPERATURE":
			data.WaterTemperature.Value = sv.Center.Value
		case "WATER_FLOW_TO_PROBES":
			data.WaterFlow.Value = sv.Center.Value
		}
	}

	tmpl, err := template.ParseFiles("web/templates/measurements.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
