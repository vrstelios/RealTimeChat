package model

import "time"

type Message struct {
	Id        string    `json:"id"`
	Room      string    `json:"room"`
	Name      string    `json:"name"`
	Message   string    `json:"message"`
	Role      string    `json:"role"`
	Timestamp time.Time `json:"timestamp"`
}
