package models

import (
	"cobranca-bmp/config"
	"fmt"
	"time"
)

type WebhookDigitacaoPayload struct {
	Url                  string `json:"url"`
	Convenio             int    `json:"convenio"`
	NumeroAcompanhamento string `json:"numero_acompanhamento"`
	IdProposta           int    `json:"id_proposta"`
}

type WebhookSimulacaoPayload struct {
	IdSimulacao       int    `json:"idSimulacao"`
	IdSimulacaoTabela int    `json:"idSimulacaoTabela"`
	IdTabela          int    `json:"idTabela"`
	Status            string `json:"status"`
	Payload           any    `json:"payload,omitempty"`
}

type WebhookElegibilidadePayload struct {
	CPF      string                  `json:"cpf"`
	Parcial  bool                    `json:"parcial,omitempty"`
	Mensagem string                  `json:"mensagem,omitempty"`
	Vinculos []VinculoWebhookPayload `json:"vinculos,omitempty"`
}

type WebhookUpdateStatus struct {
	IdProposta int `json:"idProposta"`
	Status     int `json:"status"`
	Response   any `json:"response,omitempty"`
}

type VinculoWebhookPayload struct {
	DocumentoEmpregador string  `json:"documentoEmpregador"`
	Matricula           string  `json:"matricula"`
	MargemBase          float64 `json:"margemBase"`
	MargemDisponivel    float64 `json:"margemDisponivel"`
}

type WebhookTaskData struct {
	Data    any           `json:"data"`
	Retries int64         `json:"retries"`
	Delay   time.Duration `json:"delay"`
	Url     string        `json:"url"`
	Context string        `json:"context"`
}

func NewWebhookTaskData(url string, data any, context string) WebhookTaskData {
	return WebhookTaskData{
		Data:    data,
		Url:     url,
		Retries: config.WEBHOOK_RETRIES,
		Delay:   0,
		Context: context,
	}

}

func (w *WebhookTaskData) SetTry() {
	w.Delay += config.WEBHOOK_DELAY
	w.Retries--

}

func GenerateCCBURL(numeroAcompanhamento, codParametro string) string {
	return fmt.Sprintf("%s/%s?impressao=S&tipo=ccb&code=%s&Integracao=%s&copias=1", config.CCB_URL, "Imprimir", numeroAcompanhamento, codParametro)
}

func GenerateCCBCreditoTrabalhadorURL(numeroAcompanhamento, codParametro string) string {
	return fmt.Sprintf("%s/ImprimirCCB?Code=%s&integracao=%s", config.CCB_URL, numeroAcompanhamento, codParametro)
}
