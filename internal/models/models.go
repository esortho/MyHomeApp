package models

import "time"

type PoolStatus struct {
	FlowState bool      `json:"flowState"`
	UpdatedAt time.Time `json:"updatedAt"`
	Waterflow float64   `json:"waterflow"`
}

type HueLight struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	State struct {
		On          bool   `json:"on"`
		Brightness  int    `json:"bri"`
		ColorMode   string `json:"colormode"`
		Temperature int    `json:"ct"`
		Reachable   bool   `json:"reachable"`
	} `json:"state"`
	Type      string `json:"type"`
	ModelID   string `json:"modelid"`
	UniqueID  string `json:"uniqueid"`
	ProductID string `json:"productid"`
}

type DashboardData struct {
	PoolStatus PoolStatus `json:"poolStatus"`
	Lights     []HueLight `json:"lights"`
}
