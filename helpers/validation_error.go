package helpers

import (
	"fmt"
)

type Cache interface {
	GetError(message string, field string, context string) string
}

type FieldErr map[string][]string

/*
	Estrutura de erro retornada pelo validator de structs.

Retorna erros em um padrão semelhante aos do validator do Laravel.
"Counter" não é serializado, pois a quantidade de erros é informada no cabeçalho.
*/
type ErrValidation struct {
	Message string   `json:"message"`
	Errs    FieldErr `json:"errors"`
	Counter uint     `json:"-"`
	Codes   []int    `json:"cMensagem"`
}

// Atualiza o cabeçalho da mensagem com a quantidade de erros encontrados
func (e *ErrValidation) setHeader() {
	e.Message = fmt.Sprintf("%d erros de validação:", e.Counter)

}

// Instancia um novo ErrVAlidation
func NewErrValidation() *ErrValidation {
	errs := make(FieldErr)
	codes := make([]int, 0)
	var err = &ErrValidation{Errs: errs, Codes: codes}
	err.setHeader()
	return err

}

// Adiciona vários campos de erro.
func (e *ErrValidation) AddFields(fields map[string]string) {

	for field, err := range fields {
		e.Errs[field] = append(e.Errs[field], err)
		e.Counter++

	}
	e.setHeader()

}

// Adiciona um campo de erro.
func (e *ErrValidation) AddField(field, err string) {

	e.Errs[field] = append(e.Errs[field], err)
	e.Counter++
	e.setHeader()
}

// Implementa "error".
func (e *ErrValidation) Error() string {
	return "validate error"
}

func (e *ErrValidation) GetAllMessages() []string {
	if e.Errs == nil {
		return nil
	}
	messages := make([]string, 0)
	for _, v := range e.Errs {
		messages = append(messages, v...)

	}
	return messages

}

type JRPCResponse struct {
	Version string    `json:"jsonrpc"`
	ID      any       `json:"id"`
	Err     JRPCError `json:"error"`
}

type JRPCResponses []JRPCResponse

type JRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Field   string `json:"-"`
}

func (j JRPCResponse) Error() string {
	return j.Err.Message
}

func (j JRPCResponses) Error() string {
	return ""
}

func NewJRPCResponse(cache Cache, field, message string, id any) JRPCResponse {
	//	c := cache.GetError(field, "codigo", "Global")
	//code, _ := strconv.Atoi(c)

	return JRPCResponse{
		Err: JRPCError{
			Code:    422,
			Message: message,
		},
		Version: "2.0",
		ID:      id,
	}

}
