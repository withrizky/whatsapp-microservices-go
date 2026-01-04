package model

type WaPayload struct {
	To      string `json:"to" binding:"required"`
	Message string `json:"message" binding:"required"`
}
