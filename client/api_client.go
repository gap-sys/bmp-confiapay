package client

import (
	"bytes"
	"cobranca-bmp/auth"
	"cobranca-bmp/cache"
	"cobranca-bmp/config"
	"cobranca-bmp/helpers"
	"cobranca-bmp/models"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
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

func (a APIClient) Auth(id, client string, expire bool, privateKey string, redisKey string) (string, error) {
	if !expire {
		token, err := a.cache.GetToken(redisKey)
		if err == nil && token != "" {
			return "Bearer " + token, nil
		}
	}

	jwt, err := auth.GenerateToken(id, client, config.JWT_AUD, privateKey)
	if err != nil {
		helpers.LogError(a.ctx, a.logger, a.loc, "api client", "", "Erro ao gerar token jwt", err.Error(), nil)
		return "", err
	}

	addr := os.Getenv("AUTH_URL")
	body := url.Values{}
	body.Set("client_id", "client_id")
	body.Set("grant_type", "client_credentials")
	//body.Set("client_id", client)
	body.Set("scope", "bmp.digital.api.full.access")
	body.Set("client_assertion", jwt)
	body.Set("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
	reqBody := body.Encode()

	req, err := http.NewRequest(http.MethodPost, addr, strings.NewReader(reqBody))
	if err != nil {
		helpers.LogError(a.ctx, a.logger, a.loc, "api client", "", "Erro ao criar requisição de autenticação", err.Error(), nil)
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cache-Control", "no-cache")
	//helpers.LogInfo(a.ctx, a.logger, a.loc, "api client", "api client", "Requisitando para: "+addr, nil)
	resp, err := a.doRequest(req)
	if err != nil {
		helpers.LogError(a.ctx, a.logger, a.loc, "api client", "", "Erro ao realizar requisição de autenticação", err.Error(), nil)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		var data models.APIError
		if err := a.deserialize(resp.Body, &data); err != nil {
			return "", err

		}
		if data.Msg == "" {
			data.Msg = "erro ao se autenticar na API do BMP"
		}

		return "", data
	}

	var data map[string]any

	if err := a.deserialize(resp.Body, &data); err != nil {
		return "", err

	}

	token, ok := data["access_token"].(string)
	if !ok || token == "" {
		helpers.LogError(a.ctx, a.logger, a.loc, "api client", "", "Token de acesso não encontrado", "", nil)
		return "", errors.New("token de acesso não encontrado")
	}
	expDate := time.Now().In(a.loc).Add(config.REDIS_TOKEN_EXP)
	expDateStr := expDate.String()

	err = a.cache.SetTokenExp(redisKey, token, expDateStr, expDate)
	if err != nil {
		helpers.LogError(a.ctx, a.logger, a.loc, "api client", "", "Erro ao salvar token no redis", err.Error(), nil)
	}

	return "Bearer " + token, nil
}
