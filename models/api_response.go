package models

import "strconv"

type APIResponse struct {
	ID       int       `json:"id"`
	Msg      string    `json:"msg"`
	HasError bool      `json:"hasError"`
	Messages []Message `json:"messages"`
	Codigo   string    `json:"codigo"`
	//Parametros []Parametros `json:"parametros,omitempty"`
}

type Message struct {
	MessageType int     `json:"messageType"`
	Code        string  `json:"code"`
	Context     string  `json:"context"`
	Description string  `json:"description"`
	Field       *string `json:"field"`
}

func NewAPIResponse(code, message, codigoOperacao string, contexto ...string) APIResponse {
	id, _ := strconv.Atoi(codigoOperacao)
	var msg = Message{
		Code:        code,
		Description: message,
	}

	if len(contexto) == 1 {
		msg.Context = contexto[0]
	}

	return APIResponse{
		ID:       id,
		HasError: false,
		Msg:      message,
		Messages: []Message{msg},
	}
}
