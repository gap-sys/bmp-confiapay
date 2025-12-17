package models

import "time"

type DLQData struct {
	Payload  any       `json:"payload"`
	Contexto string    `json:"contexto"`
	Mensagem string    `json:"mensagem"`
	Erro     string    `json:"erro"`
	Time     time.Time `json:"time"`
}
