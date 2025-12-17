package models

import (
	"strconv"
)

type APIError struct {
	ID       int          `json:"id,omitzero"`
	HasError bool         `json:"hasError"`
	Messages []APIMessage `json:"messages"`
	Msg      string       `json:"msg"`
	Result   any          `json:"result"`
}

type APIMessage struct {
	Code        string `json:"code,omitempty"`
	Context     string `json:"context"`
	Description string `json:"description,omitempty"`
	Field       string `json:"field,omitempty"`
	MessageType int    `json:"messageType,omitempty"`
}

func (a APIError) Error() string {
	//var rawErr json.RawMessage
	//rawErr, _ := json.Marshal(a)
	return a.Msg
}

type APIMessages []APIMessage

func (a APIMessages) Error() string {
	var msg string
	for _, m := range a {
		msg += m.Description + "  \n"
	}
	return msg
}

func NewAPIError(code, message, codigoOperacao string, contexto ...string) APIError {
	id, _ := strconv.Atoi(codigoOperacao)
	var msg = APIMessage{
		Code:        code,
		Description: message,
	}

	if len(contexto) == 1 {
		msg.Context = contexto[0]
	}

	return APIError{
		ID:       id,
		HasError: true,
		Msg:      message,
		Messages: []APIMessage{msg},
	}
}
