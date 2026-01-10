package client

import (
	"cobranca-bmp/cache"
	"cobranca-bmp/config"
	"cobranca-bmp/models"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// Representa um cliente que realiza operações referentes a propostas de empréstimo na API externa.É composto por APICLient.
type CobrancaClient struct {
	APIClient
}

func NewCobrancaClient(ctx context.Context, loc *time.Location, cache *cache.RedisCache, logger *slog.Logger, baseUrl, authUrl string) *CobrancaClient {
	return &CobrancaClient{NewApiClient(ctx, loc, cache, logger, baseUrl, authUrl)}

}

func (c *CobrancaClient) GerarCobrancaParcela(payload models.CobrancaUnicaInput, token string, idempotencyKey string) (models.GerarCobrancaResponse, int, string, error) {
	url := fmt.Sprintf("%s/%s", c.baseUrl, "AgendaRecebivel/GerarUnicaCobranca")

	req, err := c.newRequest(http.MethodPost, url, payload)
	if err != nil {
		return models.GerarCobrancaResponse{}, 500, config.API_STATUS_ERR, models.NewAPIError(config.ERR_INTERNAL_SERVER_STR, err.Error(), "")
	}
	req.Header.Set("Authorization", token)
	req.Header.Set("IdempotencyKey", idempotencyKey)

	resp, err := c.doRequest(req)
	if err != nil {
		return models.GerarCobrancaResponse{}, 500, config.API_STATUS_ERR, models.NewAPIError(config.ERR_INTERNAL_SERVER_STR, err.Error(), "")
	}
	defer resp.Body.Close()

	rawBody, _ := io.ReadAll(resp.Body)
	var APIErr models.APIError
	json.Unmarshal(rawBody, &APIErr)
	if resp.StatusCode != 200 {
		var status string
		switch resp.StatusCode {
		case http.StatusGatewayTimeout, http.StatusRequestTimeout:
			status = config.API_STATUS_TIMEOUT
		case http.StatusTooManyRequests:
			status = config.API_STATUS_RATE_LIMIT
		case http.StatusUnauthorized:
			status = config.API_STATUS_UNAUTHORIZED
			APIErr = models.NewAPIError("", "Não foi possível se autenticar no BMP", payload.Dto.CodigoOperacao)
		default:
			status = config.API_STATUS_ERR

		}

		return models.GerarCobrancaResponse{}, resp.StatusCode, status, APIErr
	}

	var data any
	if err := json.Unmarshal(rawBody, &data); err != nil {
		return models.GerarCobrancaResponse{}, 500, config.API_STATUS_ERR, models.NewAPIError(config.ERR_INTERNAL_SERVER_STR, err.Error(), "")
	}

	return models.GerarCobrancaResponse{}, resp.StatusCode, "", nil

}

func (c *CobrancaClient) GerarCobrancaParcelasMultiplas(payload models.MultiplasCobrancasInput, token string, idempotencyKey string) (models.GerarCobrancaResponse, int, string, error) {

	url := fmt.Sprintf("%s/%s", c.baseUrl, "AgendaRecebivel/GerarMultiplasCobrancas")

	req, err := c.newRequest(http.MethodPost, url, payload)
	if err != nil {
		return models.GerarCobrancaResponse{}, 500, config.API_STATUS_ERR, models.NewAPIError(config.ERR_INTERNAL_SERVER_STR, err.Error(), "")
	}
	req.Header.Set("Authorization", token)
	req.Header.Set("IdempotencyKey", idempotencyKey)

	resp, err := c.doRequest(req)
	if err != nil {
		return models.GerarCobrancaResponse{}, 500, config.API_STATUS_ERR, models.NewAPIError(config.ERR_INTERNAL_SERVER_STR, err.Error(), "")
	}
	defer resp.Body.Close()

	rawBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		var status string
		var APIErr models.APIError
		json.Unmarshal(rawBody, &APIErr)
		APIErr.Result = json.RawMessage(rawBody)
		switch resp.StatusCode {
		case http.StatusGatewayTimeout, http.StatusRequestTimeout:
			status = config.API_STATUS_TIMEOUT
			APIErr.Msg = "Geração de cobranças indisponível"
		case http.StatusTooManyRequests:
			status = config.API_STATUS_RATE_LIMIT
			APIErr.Msg = "Geração de cobranças indisponível"
		case http.StatusUnauthorized:
			status = config.API_STATUS_UNAUTHORIZED
			APIErr = models.NewAPIError("", "Não foi possível se autenticar no BMP", payload.Dto.CodigoOperacao)
			APIErr.Result = APIErr

		default:
			status = config.API_STATUS_ERR

		}

		return models.GerarCobrancaResponse{}, resp.StatusCode, status, APIErr
	}

	var data models.GerarCobrancaResponse
	if err := json.Unmarshal(rawBody, &data); err != nil {
		return models.GerarCobrancaResponse{}, 500, config.API_STATUS_ERR, models.NewAPIError(config.ERR_INTERNAL_SERVER_STR, err.Error(), "")
	}

	return data, 200, "", nil

}

func (c *CobrancaClient) CancelarCobranca(payload models.CancelarCobrancaInput, token string, idempotencyKey string) (any, int, string, error) {
	url := fmt.Sprintf("%s/%s", c.baseUrl, "AgendaRecebivel/CancelarCobrancas")

	req, err := c.newRequest(http.MethodPost, url, payload)
	if err != nil {
		return nil, 500, config.API_STATUS_ERR, models.NewAPIError(config.ERR_INTERNAL_SERVER_STR, err.Error(), "")
	}
	req.Header.Set("Authorization", token)
	req.Header.Set("IdempotencyKey", idempotencyKey)

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, 500, config.API_STATUS_ERR, models.NewAPIError(config.ERR_INTERNAL_SERVER_STR, err.Error(), "")
	}
	defer resp.Body.Close()

	rawBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		var status string
		var APIErr models.APIError
		json.Unmarshal(rawBody, &APIErr)

		switch resp.StatusCode {
		case http.StatusGatewayTimeout, http.StatusRequestTimeout:
			status = config.API_STATUS_TIMEOUT
		case http.StatusTooManyRequests:
			status = config.API_STATUS_RATE_LIMIT
		case http.StatusUnauthorized:
			status = config.API_STATUS_UNAUTHORIZED
			APIErr = models.NewAPIError("", "Não foi possível se autenticar no BMP", payload.DTO.CodigoOperacao)

		default:
			status = config.API_STATUS_ERR

		}
		return nil, resp.StatusCode, status, APIErr
	}

	var data any
	if err := json.Unmarshal(rawBody, &data); err != nil {
		return models.ConsultaCobrancaResponse{}, 500, config.API_STATUS_ERR, models.NewAPIError(config.ERR_INTERNAL_SERVER_STR, err.Error(), "")
	}

	return data, 200, "", nil

}

func (c *CobrancaClient) LancamentoParcela(payload models.LancamentoParcelaAPIInput, token string, idempotencyKey string) (any, int, string, error) {
	url := fmt.Sprintf("%s/%s", c.baseUrl, "AgendaRecebivel/IncluirLancamentoParcela")

	req, err := c.newRequest(http.MethodPost, url, payload)
	if err != nil {
		return nil, 500, config.API_STATUS_ERR, models.NewAPIError(config.ERR_INTERNAL_SERVER_STR, err.Error(), "")
	}
	req.Header.Set("Authorization", token)
	req.Header.Set("IdempotencyKey", idempotencyKey)

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, 500, config.API_STATUS_ERR, models.NewAPIError(config.ERR_INTERNAL_SERVER_STR, err.Error(), "")
	}
	defer resp.Body.Close()

	rawBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		var status string
		var APIErr models.APIError
		json.Unmarshal(rawBody, &APIErr)

		switch resp.StatusCode {
		case http.StatusGatewayTimeout, http.StatusRequestTimeout:
			status = config.API_STATUS_TIMEOUT
		case http.StatusTooManyRequests:
			status = config.API_STATUS_RATE_LIMIT
		case http.StatusUnauthorized:
			status = config.API_STATUS_UNAUTHORIZED
			APIErr = models.NewAPIError("", "Não foi possível se autenticar no BMP", payload.DTO.CodigoOperacao)

		default:
			status = config.API_STATUS_ERR

		}
		return nil, resp.StatusCode, status, APIErr
	}

	var data any
	if err := json.Unmarshal(rawBody, &data); err != nil {
		return models.ConsultaCobrancaResponse{}, 500, config.API_STATUS_ERR, models.NewAPIError(config.ERR_INTERNAL_SERVER_STR, err.Error(), "")
	}

	return data, 200, "", nil

}

func (c *CobrancaClient) ConsultarCobranca(payload models.ConsultarDetalhesInput, token string, idempotencyKey string) (models.ConsultaCobrancaResponse, int, string, error) {
	url := fmt.Sprintf("%s/%s", c.baseUrl, "AgendaRecebivel/ConsultarDetalhes")

	req, err := c.newRequest(http.MethodPost, url, payload)
	if err != nil {
		return models.ConsultaCobrancaResponse{}, 500, config.API_STATUS_ERR, models.NewAPIError(config.ERR_INTERNAL_SERVER_STR, err.Error(), "")
	}
	req.Header.Set("Authorization", token)
	req.Header.Set("IdempotencyKey", idempotencyKey)

	resp, err := c.doRequest(req)
	if err != nil {
		return models.ConsultaCobrancaResponse{}, 500, config.API_STATUS_ERR, models.NewAPIError(config.ERR_INTERNAL_SERVER_STR, err.Error(), "")
	}
	defer resp.Body.Close()

	rawBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		var APIErr models.APIError
		json.Unmarshal(rawBody, &APIErr)
		var status string
		switch resp.StatusCode {
		case http.StatusGatewayTimeout, http.StatusRequestTimeout:
			status = config.API_STATUS_TIMEOUT
		case http.StatusTooManyRequests:
			status = config.API_STATUS_RATE_LIMIT
		case http.StatusUnauthorized:
			status = config.API_STATUS_UNAUTHORIZED
			APIErr = models.NewAPIError("", "Não foi possível se autenticar no BMP", payload.DTO.CodigoOperacao)

		default:
			status = config.API_STATUS_ERR

		}

		return models.ConsultaCobrancaResponse{}, resp.StatusCode, status, APIErr
	}

	var data models.ConsultaCobrancaResponse
	if err := json.Unmarshal(rawBody, &data); err != nil {
		return models.ConsultaCobrancaResponse{}, 500, config.API_STATUS_ERR, models.NewAPIError(config.ERR_INTERNAL_SERVER_STR, err.Error(), "")
	}
	var result = make(map[string]any)

	if err = json.Unmarshal(rawBody, &result); err != nil {
		return models.ConsultaCobrancaResponse{}, 500, config.API_STATUS_ERR, models.NewAPIError(config.ERR_INTERNAL_SERVER_STR, err.Error(), "")
	}

	data.Result = result

	return data, 200, "", nil

}

func (c *CobrancaClient) ConsultarBoleto(payload models.ConsultaBoletoInput, token string, idempotencyKey string) (models.BoletoConsultado, int, string, error) {
	url := fmt.Sprintf("%s/%s", c.baseUrl, "AgendaRecebivel/ConsultarBoletos")

	req, err := c.newRequest(http.MethodPost, url, payload)
	if err != nil {
		return models.BoletoConsultado{}, 500, config.API_STATUS_ERR, models.NewAPIError(config.ERR_INTERNAL_SERVER_STR, err.Error(), "")
	}
	req.Header.Set("Authorization", token)
	req.Header.Set("IdempotencyKey", idempotencyKey)

	resp, err := c.doRequest(req)
	if err != nil {
		return models.BoletoConsultado{}, 500, config.API_STATUS_ERR, models.NewAPIError(config.ERR_INTERNAL_SERVER_STR, err.Error(), "")
	}
	defer resp.Body.Close()

	rawBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		var APIErr models.APIError
		json.Unmarshal(rawBody, &APIErr)
		var status string
		switch resp.StatusCode {
		case http.StatusGatewayTimeout, http.StatusRequestTimeout:
			status = config.API_STATUS_TIMEOUT
		case http.StatusTooManyRequests:
			status = config.API_STATUS_RATE_LIMIT
		case http.StatusUnauthorized:
			status = config.API_STATUS_UNAUTHORIZED
			APIErr = models.NewAPIError("", "Não foi possível se autenticar no BMP", payload.DTO.CodigoOperacao)

		default:
			status = config.API_STATUS_ERR

		}

		return models.BoletoConsultado{}, resp.StatusCode, status, APIErr
	}

	var data models.BoletoConsultado
	if err := json.Unmarshal(rawBody, &data); err != nil {
		return models.BoletoConsultado{}, 500, config.API_STATUS_ERR, models.NewAPIError(config.ERR_INTERNAL_SERVER_STR, err.Error(), "")
	}

	/*var result = make(map[string]any)

	if err = json.Unmarshal(rawBody, &result); err != nil {
		return models.BoletoConsultado{}, 500, config.API_STATUS_ERR, models.NewAPIError(config.ERR_INTERNAL_SERVER_STR, err.Error(), "")
	}*/

	//data.Result = result

	return data, 200, "", nil

}

// Busca propostas na API do BMP por ID ou código de proposta.
func (c *CobrancaClient) BuscarProposta(param models.BuscarProposta, token, idempotencyKey string) (models.PropostaAPIResponse, string, int, error) {
	url := fmt.Sprintf("%s/%s", c.baseUrl, "Proposta/Consultar")
	var dto = models.DTO{Dto: param}
	req, err := c.newRequest(http.MethodPost, url, dto)
	if err != nil {
		return models.PropostaAPIResponse{}, config.API_STATUS_ERR, 500, models.NewAPIError(config.ERR_INTERNAL_SERVER_STR, err.Error(), param.CodigoOperacao)
	}

	req.Header.Set("Authorization", token)
	req.Header.Set("IdempotencyKey", idempotencyKey)

	resp, err := c.doRequest(req)
	if err != nil {
		return models.PropostaAPIResponse{}, config.API_STATUS_ERR, 500, models.NewAPIError(config.ERR_INTERNAL_SERVER_STR, err.Error(), param.CodigoOperacao)
	}
	defer resp.Body.Close()

	var rawData any
	rawBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(rawBody, &rawData)

	if resp.StatusCode != 200 {
		if resp.StatusCode == 401 {
			return models.PropostaAPIResponse{}, config.API_STATUS_ERR, 401, models.NewAPIError("", "Erro ao se autenticar na API do BMP", param.CodigoProposta, "CONFIAPAY")

		}
		var APIErr models.APIError
		if err := json.Unmarshal(rawBody, &APIErr); err != nil {
			return models.PropostaAPIResponse{}, config.API_STATUS_ERR, 500, models.NewAPIError("", err.Error(), param.CodigoOperacao)

		}
		return models.PropostaAPIResponse{}, config.API_STATUS_ERR, resp.StatusCode, APIErr
	}

	var data models.PropostaAPIResponse
	if err := json.Unmarshal(rawBody, &data); err != nil {
		return models.PropostaAPIResponse{}, config.API_STATUS_ERR, 500, models.NewAPIError("", err.Error(), param.CodigoOperacao)
	}
	data.Result = rawData

	return data, "", 200, nil

}
