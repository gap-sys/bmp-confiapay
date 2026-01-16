package client

import (
	"bytes"
	"cobranca-bmp/cache"
	"cobranca-bmp/helpers"
	"cobranca-bmp/models"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// Api client representa um cliente para a API externa(será extendido para outros clientes específicos neste pacote)
type APIClient struct {
	client  *http.Client
	cache   *cache.RedisCache
	baseUrl string
	authUrl string
	loc     *time.Location
	ctx     context.Context
	logger  *slog.Logger
}

func NewApiClient(ctx context.Context, loc *time.Location, cache *cache.RedisCache, logger *slog.Logger, baseUrl, authUrl string) APIClient {
	return APIClient{
		ctx:     ctx,
		client:  &http.Client{},
		cache:   cache,
		baseUrl: baseUrl,
		authUrl: authUrl,
		logger:  logger,
		loc:     loc,
	}
}

// Cria uma nova requisição com o método e a url passados e um io.Reader com data serializada em json.
func (a APIClient) newRequest(method, url string, data any) (*http.Request, error) {
	var body io.Reader
	var err error
	var json []byte

	if data == nil {
		body = nil
	} else {
		json, err = a.serialize(data)
		if err != nil {
			return nil, err
		}

		body = bytes.NewBuffer(json)
	}

	req, err := http.NewRequest(method, url, body)

	if err != nil {
		helpers.LogError(a.ctx, a.logger, a.loc, "api client", "", "Erro ao criar requisição para:"+url, err.Error(), data)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	//helpers.LogInfo(a.ctx, a.logger, a.loc, "api client", "", "Requisitando para: "+url, data)
	return req, err

}

// Realiza uma requisição utilizando o http.Cliente do EmprestimoCliente e retorna a resposta ou possíveis erros.
func (a APIClient) doRequest(req *http.Request) (*http.Response, error) {
	curl, _ := helpers.GetCurlCommand(req)
	helpers.LogInfo(a.ctx, a.logger, a.loc, "api client", "", "Requisitando para:"+req.URL.String(), map[string]string{"curl": curl.String()})
	resp, err := a.client.Do(req)
	if err != nil {
		helpers.LogError(a.ctx, a.logger, a.loc, "api client", "", "Erro ao requisitar para:"+req.URL.String(), err.Error(), nil)
		return nil, err
	}
	return resp, nil
}

// Retorna um io.Reader com o json dos dados a serem utilizados como payload
func (a APIClient) serialize(data any) ([]byte, error) {
	json, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return json, nil
}

// Deserializa o json de src(que será sempre o corpo da requisição) em dst. dst deve ser sempre um ponteiro caso a intenção seja utilizar structs.
func (a APIClient) deserialize(src io.Reader, dst any) error {
	if err := json.NewDecoder(src).Decode(dst); err != nil {
		helpers.LogError(a.ctx, a.logger, a.loc, "api client", "", "Erro ao desserializar", err.Error(), nil)
		return err
	}
	return nil
}

func (a APIClient) Auth(data models.AuthPayload, expire bool) (string, error) {
	var cred = map[string]any{
		"idConvenio":       data.IdConvenio,
		"idSecuritizadora": data.IdSecuritizadora,
		"expire":           expire,
	}

	req, err := a.newRequest("POST", a.authUrl, cred)
	if err != nil {
		APIErr := models.NewAPIError("", "Falha na autenticação com o BMP. Caso o erro persista, entre em contato com o suporte.", data.Id)
		APIErr.Result = err.Error()
		return "", APIErr
	}

	resp, err := a.doRequest(req)
	if err != nil {
		APIErr := models.NewAPIError("", "Falha na autenticação com o BMP. Caso o erro persista, entre em contato com o suporte.", data.Id)
		APIErr.Result = err.Error()
		return "", APIErr
	}

	if resp.StatusCode != 200 {
		var apiErr models.APIError
		if err := a.deserialize(resp.Body, &apiErr); err != nil {
			APIErr := models.NewAPIError("", "Falha na autenticação com o BMP. Caso o erro persista, entre em contato com o suporte.", data.Id)
			APIErr.Result = err.Error()
			return "", APIErr

		}

		apiErr.HasError = true
		if apiErr.Msg == "" {
			apiErr.Msg = "Falha na autenticação com o BMP. Caso o erro persista, entre em contato com o suporte."
		}
		return "", apiErr

	}

	var respBody map[string]string

	if err := a.deserialize(resp.Body, &respBody); err != nil {
		return "", models.NewAPIError("", err.Error(), data.Id)

	}

	token, ok := respBody["token"]
	if !ok {
		APIErr := models.NewAPIError("", "Falha na autenticação com o BMP. Caso o erro persista, entre em contato com o suporte.", data.Id)
		APIErr.Result = respBody
		return "", APIErr
	}

	return "BEARER " + token, nil

}
