package handlers

import (
	"cobranca-bmp/config"
	"cobranca-bmp/helpers"
	"cobranca-bmp/models"
	"time"

	"github.com/gofiber/fiber/v2"
)

type QOSSetter interface {
	SetQOS(qos int) error
	GetQOS() map[string]any
}

type ConfigController struct {
	qosSetter   QOSSetter
	dbManager   DBManager
	poolReseter PoolReseter
}

type PoolReseter interface {
	SetInterval(interval time.Duration)
}

type DBManager interface {
	Reconfig(config *config.DBPoolConfig, reconnect bool) any
}

func NewConfigController(qosSetter QOSSetter, dbManager DBManager, poolReseter PoolReseter) *ConfigController {
	return &ConfigController{
		qosSetter:   qosSetter,
		dbManager:   dbManager,
		poolReseter: poolReseter}
}

func (cc *ConfigController) GetPrefix() string {
	return "config"
}

func (cc *ConfigController) Route(r fiber.Router) {
	r.Post("/qos", cc.SetQOS())
	r.Post("/db/set", cc.ReconfigDB())
	r.Post(("/env"), cc.SetGlobalEnvVars())

}

func (cc *ConfigController) SetQOS() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var settings models.SetQOS
		if err := c.BodyParser(&settings); err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(err.Error())
		}

		if err := cc.qosSetter.SetQOS(settings.QOS); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(err)

		}

		return c.JSON(cc.qosSetter.GetQOS())

	}

}

// ReconfigDB godoc
//
//	@Summary		Reconfigurar DB.
//	@Description	Reconfigura os valores do DB.
//	@Tags			Config
//	@Accept			json
//	@Produce		json
//	@Param			body	body        config.APIDBPoolConfig	true  "Parâmetros do banco de dados."
//	@Success		200		{object}	map[string]any
//	@Failure		422		{object}	models.JRPCResponse
//	@Router			/config/db/set [post]
func (cc *ConfigController) ReconfigDB() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input config.APIDBPoolConfig
		if err := c.BodyParser(&input); err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(err.Error())
		}

		var configDB *config.DBPoolConfig
		if input.Default {
			configDB = config.NewDBPoolConfigFromEnv()

		} else {
			configDB = &config.DBPoolConfig{
				MaxOpenConns:    input.MaxOpenConns,
				MaxIdleConns:    input.MaxIdleConns,
				ConnMaxLifetime: time.Duration(input.ConnMaxLifetime) * time.Second,
				ConnMaxIdleTime: time.Duration(input.ConnMaxIdleTime) * time.Second,
			}
		}

		data := cc.dbManager.Reconfig(configDB, input.Reconnect)

		return c.JSON(data)

	}

}

// SetGlobalEnvVars godoc
//
//	@Summary		Configurar variáveis de ambiente globais.
//	@Description	Configura variáveis de ambiente.
//	@Tags			Config
//	@Accept			json
//	@Produce		json
//	@Param			body	body	config.GlobalEnvVars	true  "Variáveis que se deseja alterar."
//	@Success		200		{object}	map[string]any
//	@Failure		422		{object}	models.JRPCResponse
//	@Router			/config/env [post]
func (cc *ConfigController) SetGlobalEnvVars() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input config.GlobalEnvVars
		if err := c.BodyParser(&input); err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": helpers.ParseJsonError(err.Error())})
		}

		data := config.SetEnvVars(input)
		if input.DbMonitoringInterval > 0 {
			cc.poolReseter.SetInterval(time.Duration(input.DbMonitoringInterval) * time.Second)
		}
		return c.JSON(data)

	}

}
