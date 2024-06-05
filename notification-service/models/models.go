package models

type Notification struct {
	ID      int32  `json:"id"`
	Message string `json:"message"`
	Channel string `json:"channel"`
	Status  string `json:"status"`
}
