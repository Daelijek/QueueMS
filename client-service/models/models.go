package models

type Queue struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

type Client struct {
	ID      int32  `json:"id"`
	QueueID int32  `json:"queue_id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
}
