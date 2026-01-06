package config

import (
	"fmt"
	"strings"
	"sync"

	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

var globalMu sync.Mutex
var HOMOLOG_LOCAL = false

// Representa variáveis de ambiente globais que podem ser alteradas em runtime.
type EnvVars struct {
	TimeoutRetries         int           `json:"timeoutRetries"`          //Número de tentativas em caso de timeout
	TimeoutDelay           int           `json:"timeoutDelay"`            //Delay entre uma tentativa e outra na fila,em caso de timeout
	DbDelay                int           `json:"dbDelay"`                 //Delay entre uma tentativa e outra na fila, em caso de falha ao atualizar o banco de dados
	InitialDelay           int           `json:"initialDelay"`            //Delay inicial de uma simulação em fila.
	RateLimitRetries       int           `json:"rateLimitRetries"`        //Número de tentativas em caso de rate limit
	RateLimitDelay         int           `json:"rateLimitDelay"`          //Delay entre uma tentativa e outra na fila,em caso de rate limit
	WebhookRetries         int           `json:"webhookRetries"`          //Número de tentativas em caso de falha no webhook
	WebhookDelay           int           `json:"webhookDelay"`            //Delay entre uma tentativa e outra na fila,em caso de falha no webhook
	WebhookKey             string        `json:"webhookKey"`              //Chave de autenticação do webhook do Confiapay
	WebhookHash            string        `json:"webhookHash"`             //Hash passado no webhook do Confiapay
	WebhookUrl             string        `json:"webhookUrl"`              //URL do webhook do Confiapay
	DbMonitoringInterval   int           `json:"dbMonitoringInterval"`    //Intervalo de tempo em segundos que o DB será monitorado.
	DbQueryTimeout         int           `json:"dbQueryTimeout"`          //Timeout em segundos de cada operação de uma transaction no db.
	DbMonitoringMaxAlerts  int           `json:"dbMonitoringMaxAlerts"`   //Número máximo de alertas guardados pelo monitor do DB.
	RedisTokenExpiration   time.Duration `json:"redisTokenExpiration"`    //Expiração de um token no Redis(em minutos)
	GerarMultipasCobrancas *bool         `json:"gerarMultiplasCobrancas"` //Define se será chamado ou não o endpoint de geração de múltiplas cobranças no BMP
	//	IOF                   float64 `json:"iof"`                   //valor do IOF

}

// Representa informações referentes a movimentações e de/para de um status do BMP recebido na callback.
type StatusInfo struct {
	Status        int
	TipoData      int
	FluxoToRemove []int
	SetFluxo      bool
	Nome          string
}

// Status de erro utilizados para controle interno das operações de digitação e simulação
const (
	API_STATUS_UNAUTHORIZED = "Não autorizado"
	API_STATUS_TIMEOUT      = "Timeout"
	API_STATUS_RATE_LIMIT   = "Status Rate Limit"
	API_STATUS_ERR          = "Erro"
)

// Variáveis do DB
var (
	DB_PORT                           int
	DB_URL                            string
	DB_MAX_OPEN_CONNS                 int
	DB_MAX_IDLE_CONNS                 int
	REDIS_TOKEN_EXP                   time.Duration
	DB_CONN_MAX_LIFETIME              time.Duration
	DB_CONN_MAX_IDLE_TIME             time.Duration
	DB_POOL_MONITORING_INTERVAL       time.Duration
	DB_QUERY_TIMEOUT                  time.Duration
	DB_POOL_MONITORING_MAX_ALERTS     int
	DB_POOL_MONITORING_ALERT_INTERVAL time.Duration
)

var (
	IPS = []string{}
)

var (
	TIPO_COBRANCA_BOLETO      = 2
	TIPO_COBRANCA_PIX         = 3
	TIPO_COBRANCA_BOLETOPIX   = 7
	STATUS_CONSULTAR_COBRANCA = "consultar_cobranca"
	STATUS_GERAR_COBRANCA     = "gerar_cobranca"
	STATUS_CANCELAR_COBRANCA  = "cancelar_cobranca"
	STATUS_LANCAMENTO_PARCELA = "lancamento_parcela"
	GERAR_MULTIPLAS_COBRANCAS = true
)

var (
	RABBITMQ_URL           string
	RABBITMQ_QOS           int
	RABBITMQ_SIMULACAO_QOS int
	RABBITMQ_GLOBAL_QOS    bool
	WEBHOOK_QUEUE          string
	DADOS_BANCARIOS_QUEUE  string
	DB_QUEUE               string
	BMP_EXCHANGE           string
	DLQ_EXCHANGE           string

	COBRANCA_QUEUE string
	DLQ_QUEUE      string

	INITIAL_DELAY          time.Duration
	TIMEOUT_DELAY          time.Duration
	RATE_LIMIT_DELAY       time.Duration
	DB_QUEUE_DELAY         time.Duration
	TIMEOUT_MAX_RETRIES    int64
	RATE_LIMIT_MAX_RETRIES int64
	DB_UPDATE_MAX_RETRIES  int64
	BASE_URL               string

	AUTH_URL   string
	REDIS_URL  string
	REDIS_DBCH string
	API_KEY    string

	WEBHOOK_KEY           string
	WEBHOOK_HASH          string
	WEBHOOK_RETRIES       int64
	WEBHOOK_DELAY         time.Duration
	WEBHOOK_DIGITACAO_URL string
	WEBHOOK_LIBERACAO_URL string
)

// Status de erro padrão JRPC
const (
	ERR_PARSE               = 32700
	ERR_INVALID_REQ         = 32600
	ERR_INVALID_METHOD      = 32601
	ERR_INVALID_PARAMS      = 32602
	ERR_INTERNAL_SERVER     = 32603
	ERR_PARSE_STR           = "32700"
	ERR_INVALID_REQ_STR     = "32600"
	ERR_INVALID_METHOD_STR  = "32601"
	ERR_INVALID_PARAMS_STR  = "32602"
	ERR_INTERNAL_SERVER_STR = "32603"
)

// Carrega as variáveis de ambiente para o sistema operacional, verificando os arquivos passados em "files".Caso nenhum arquivo seja passado, abrirá o arquivo .env que se encontra na raiz do projeto
func LoadEnvVar(files ...string) error {
	err := godotenv.Load(files...)
	if err != nil {
		return err
	}

	return InitializeEnvVar()
}

// Recupera as variáveis de ambiente setadas em LoadEnvVar e as atribui a variáveis que serão utilizadas no programa.
func InitializeEnvVar() error {
	DADOS_BANCARIOS_QUEUE = os.Getenv("DADOS_BANCARIOS_QUEUE")
	DB_QUEUE = os.Getenv("DB_QUEUE")
	BMP_EXCHANGE = os.Getenv("BMP_EXCHANGE")
	DLQ_EXCHANGE = os.Getenv("DLQ_EXCHANGE")
	DLQ_QUEUE = os.Getenv("DLQ_QUEUE")
	BASE_URL = os.Getenv("BASE_URL")
	AUTH_URL = os.Getenv("AUTH_URL")
	RABBITMQ_URL = os.Getenv("RABBITMQ_URL")
	REDIS_URL = os.Getenv("REDIS_URL")
	REDIS_DBCH = os.Getenv("REDIS_DBCH")
	API_KEY = os.Getenv("API_KEY")
	WEBHOOK_KEY = os.Getenv("WEBHOOK_KEY")
	WEBHOOK_HASH = os.Getenv("WEBHOOK_HASH")
	WEBHOOK_QUEUE = os.Getenv("WEBHOOK_QUEUE")
	WEBHOOK_LIBERACAO_URL = os.Getenv("WEBHOOK_LIBERACAO_URL")
	COBRANCA_QUEUE = os.Getenv("COBRANCA_QUEUE")

	IPS = strings.Split(os.Getenv("IPS"), ",")

	delay, err := strconv.ParseInt(os.Getenv("TIMEOUT_DELAY"), 10, 64)
	if err != nil {
		return err
	}
	TIMEOUT_DELAY = time.Duration(delay) * time.Second

	delay, err = strconv.ParseInt(os.Getenv("INITIAL_DELAY"), 10, 64)
	if err != nil {
		return err
	}

	INITIAL_DELAY = time.Duration(delay) * time.Second

	delay, err = strconv.ParseInt(os.Getenv("RATE_LIMIT_DELAY"), 10, 64)
	if err != nil {
		return err
	}

	RATE_LIMIT_DELAY = time.Duration(delay) * time.Second

	delay, err = strconv.ParseInt(os.Getenv("DB_QUEUE_DELAY"), 10, 64)
	if err != nil {
		return err
	}
	DB_QUEUE_DELAY = time.Duration(delay) * time.Second

	delay, err = strconv.ParseInt(os.Getenv("REDIS_TOKEN_EXP"), 10, 64)
	if err != nil {
		return err
	}
	REDIS_TOKEN_EXP = time.Duration(delay) * time.Minute

	TIMEOUT_MAX_RETRIES, err = strconv.ParseInt(os.Getenv("TIMEOUT_MAX_RETRIES"), 10, 64)
	if err != nil {
		return err
	}

	RATE_LIMIT_MAX_RETRIES, err = strconv.ParseInt(os.Getenv("RATE_LIMIT_MAX_RETRIES"), 10, 64)
	if err != nil {
		return err
	}

	DB_UPDATE_MAX_RETRIES, err = strconv.ParseInt(os.Getenv("DB_UPDATE_MAX_RETRIES"), 10, 64)
	if err != nil {
		return err
	}

	port, err := strconv.ParseInt(os.Getenv("DB_PORT"), 10, 64)
	if err != nil {
		return err
	}
	DB_PORT = int(port)

	WEBHOOK_RETRIES, err = strconv.ParseInt(os.Getenv("WEBHOOK_RETRIES"), 10, 64)
	if err != nil {
		return err
	}

	if HOMOLOG_LOCAL {
		setLocalHomologVars()
	}

	delay, err = strconv.ParseInt(os.Getenv("WEBHOOK_DELAY"), 10, 64)
	if err != nil {
		return err
	}
	WEBHOOK_DELAY = time.Duration(delay) * time.Second

	WEBHOOK_DIGITACAO_URL = os.Getenv("WEBHOOK_URL")

	DB_URL = fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable", os.Getenv("DB_HOST"),
		DB_PORT, os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB"))

	DB_MAX_OPEN_CONNS, err = strconv.Atoi(os.Getenv("DB_MAX_OPEN_CONNS"))
	if err != nil {
		return err
	}
	DB_MAX_IDLE_CONNS, err = strconv.Atoi(os.Getenv("DB_MAX_IDLE_CONNS"))
	if err != nil {
		return err
	}

	delay, err = strconv.ParseInt(os.Getenv("DB_CONN_MAX_LIFETIME"), 10, 64)
	if err != nil {
		return err
	}
	DB_CONN_MAX_LIFETIME = time.Duration(delay) * time.Second

	delay, err = strconv.ParseInt(os.Getenv("DB_CONN_MAX_IDLE_TIME"), 10, 64)
	if err != nil {
		return err
	}
	DB_CONN_MAX_IDLE_TIME = time.Duration(delay) * time.Second

	delay, err = strconv.ParseInt(os.Getenv("DB_QUERY_TIMEOUT"), 10, 64)
	if err != nil {
		return err
	}
	DB_QUERY_TIMEOUT = time.Duration(delay) * time.Second

	delay, err = strconv.ParseInt(os.Getenv("DB_POOL_MONITORING_INTERVAL"), 10, 64)
	if err != nil {
		return err
	}
	DB_POOL_MONITORING_INTERVAL = time.Duration(delay) * time.Second

	DB_POOL_MONITORING_MAX_ALERTS, err = strconv.Atoi(os.Getenv("DB_POOL_MONITORING_MAX_ALERTS"))
	if err != nil {
		return err
	}

	delay, err = strconv.ParseInt(os.Getenv("DB_POOL_MONITORING_ALERT_INTERVAL"), 10, 64)
	if err != nil {
		return err
	}
	DB_POOL_MONITORING_ALERT_INTERVAL = time.Duration(delay) * time.Second

	digitacaoQos, err := strconv.ParseInt(os.Getenv("RABBITMQ_DIGITACAO_QOS"), 10, 64)
	if err != nil {
		return err
	}

	RABBITMQ_QOS = int(digitacaoQos)

	simulacaooQos, err := strconv.ParseInt(os.Getenv("RABBITMQ_DIGITACAO_QOS"), 10, 64)
	if err != nil {
		return err
	}
	RABBITMQ_SIMULACAO_QOS = int(simulacaooQos)

	if HOMOLOG_LOCAL {
		setLocalHomologVars()
	}

	return nil

}

// SetEnvVars modifica as variáveis de ambiente.
func SetEnvVars(env EnvVars) map[string]any {
	globalMu.Lock()
	defer globalMu.Unlock()
	var changed = make(map[string]any)
	if env.TimeoutRetries > 0 {
		TIMEOUT_MAX_RETRIES = int64(env.TimeoutRetries)
		changed["TimeoutRetries"] = TIMEOUT_MAX_RETRIES
	}
	if env.RateLimitRetries > 0 {
		RATE_LIMIT_MAX_RETRIES = int64(env.RateLimitRetries)
		changed["RateLimitRetries"] = RATE_LIMIT_MAX_RETRIES
	}

	if env.TimeoutDelay > 0 {
		TIMEOUT_DELAY = time.Duration(env.TimeoutDelay) * time.Second
		changed["TimeoutDelay"] = TIMEOUT_DELAY.String()
	}

	if env.RateLimitDelay > 0 {
		RATE_LIMIT_DELAY = time.Duration(env.RateLimitDelay) * time.Second
		changed["RateLimitDelay"] = RATE_LIMIT_DELAY.String()
	}

	if env.DbDelay > 0 {
		DB_QUEUE_DELAY = time.Duration(env.DbDelay) * time.Second
		changed["DbDelay"] = DB_QUEUE_DELAY.String()
	}

	if env.InitialDelay > 0 {
		INITIAL_DELAY = time.Duration(env.InitialDelay) * time.Second
		changed["InitialDelay"] = INITIAL_DELAY.String()
	}

	if env.WebhookDelay > 0 {
		WEBHOOK_DELAY = time.Duration(env.WebhookDelay)
		changed["WebhookDelay"] = WEBHOOK_DELAY.String()
	}

	if env.WebhookRetries > 0 {
		WEBHOOK_RETRIES = int64(env.WebhookRetries)
		changed["WebhookRetries"] = WEBHOOK_RETRIES
	}

	if env.DbMonitoringInterval != 0 {
		DB_POOL_MONITORING_INTERVAL = time.Duration(env.DbMonitoringInterval) * time.Second
		changed["DbMonitoringInterval"] = DB_POOL_MONITORING_INTERVAL.String()

	}

	if env.DbMonitoringMaxAlerts != 0 {
		DB_POOL_MONITORING_MAX_ALERTS = env.DbMonitoringMaxAlerts
		changed["DbMonitoringMaxAlerts"] = env.DbMonitoringMaxAlerts
	}

	if env.DbQueryTimeout != 0 {
		DB_QUERY_TIMEOUT = time.Duration(env.DbQueryTimeout) * time.Second
		changed["DbQueryTimeout"] = DB_QUERY_TIMEOUT.String()

	}

	if env.RedisTokenExpiration != 0 {
		REDIS_TOKEN_EXP = time.Duration(env.RedisTokenExpiration) * time.Minute
		changed["RedisTokenExpiration"] = REDIS_TOKEN_EXP.String()

	}

	if env.GerarMultipasCobrancas != nil {
		GERAR_MULTIPLAS_COBRANCAS = *env.GerarMultipasCobrancas
		changed["GerarMultiplasCobrancas"] = GERAR_MULTIPLAS_COBRANCAS

	}

	return changed

}

func setLocalHomologVars() {
	DB_PORT = 5433

	DB_URL = fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable", os.Getenv("DB_HOST"),
		DB_PORT, os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB"))

	RABBITMQ_URL = "amqp://guest:guest@rabbitmq:5672/default"

}
