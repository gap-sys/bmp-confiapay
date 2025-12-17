package config

import (
	//"time"

	_ "cobranca-bmp/docs"
	"cobranca-bmp/helpers"
	"context"
	"encoding/json"
	"strings"

	"log"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger"
)

// Controller representa um objeto que retorna o prefixo de suas rotas e as configura internamente.
type Controller interface {
	GetPrefix() string
	Route(fiber.Router)
}

// Configura o app dividindo as rotas por prefixo e as atribuindo ao seu devido handler
func ConfigRoutes(app *fiber.App, l *slog.Logger, controllers ...Controller) {
	configMiddlewares(app, l)
	for _, controller := range controllers {
		group := app.Group(controller.GetPrefix())
		controller.Route(group)

	}

}

// Configura middlewares da aplicação.
func configMiddlewares(app *fiber.App, l *slog.Logger) {
	configLogger(app, l)
	configSwagger(app)

}

// Configura o midlleware de logger

func configLogger(app *fiber.App, l *slog.Logger) {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		log.Fatal("Erro ao carregar fuso horário para America/Sao_Paulo", err)
	}
	handler := logger.New(logger.Config{
		Output:        os.Stdout,
		Format:        "${format}", // \n${time}-${method}-${status}${red}${errorFmt} ${reset}${path}\n\t${body}\n
		TimeFormat:    "02-Jan-2006 15:04:05",
		TimeZone:      "America/Sao_Paulo",
		DisableColors: true,
		CustomTags: map[string]logger.LogFunc{
			"format": generateLogfunc(loc, l),
		},
		//Done: func(c *fiber.Ctx, logString []byte) { fmt.Println("LOGSTRING", string(logString)) },
	})

	app.Use(handler)

}

// Configura o middleware de swagger, onde será servida a documentação da API.
func configSwagger(app *fiber.App) {
	//cfg := swagger.Config{
	//FilePath: "./docs/swagger.json",
	//URL: "/docs/swagger.json",
	//}
	app.Get("/docs/*", swagger.HandlerDefault)

	//app.Use(swagger.New(cfg))
	//app.Static("docs/", "./docs/")

}

/*
func ipMiddleware(c *fiber.Ctx) error {
	fmt.Println("c.IPS:", c.IPs(), "c.IP", c.IP())
	fmt.Println("\n\n\n Headers" + c.Request().Header.String() + "\n\n\n")
	if !slices.Contains(IPS, c.IP()) {
		//return c.SendStatus(fiber.StatusForbidden)
	}
	if ip := c.Get("CF-Connecting-IP"); ip != "" {
		fmt.Println("Recebido CF-Connecting-IP", ip)

	} else {
		fmt.Println("CF-Connecting-IP vazio", ip)
	}
	return c.Next()
}
*/

func generateLogfunc(loc *time.Location, l *slog.Logger) func(output logger.Buffer, c *fiber.Ctx, data *logger.Data, extraParam string) (int, error) {

	return func(output logger.Buffer, c *fiber.Ctx, data *logger.Data, extraParam string) (int, error) {
		var logData = make(map[string]any)
		var payload = make(map[string]any)
		var path = c.Path()
		var msg string
		var logCtx string
		if strings.Contains(path, "/webhook") {
			msg = "Log Webhook"
			logCtx = "Webhook"
		} else {
			msg = "Log API"
			logCtx = "API"
		}

		var statusCode string = strconv.Itoa(c.Response().StatusCode())

		logData["message"] = msg
		//logData["method"] = c.Method()
		//logData["status"] = statusCode
		//logData["contexto"] = c.Path()
		//logData["host"] = c.Hostname()

		if data.ChainErr != nil {
			err := data.ChainErr.Error()
			logData["message"] = err
			msg = err
		}

		var reqBody any
		var body = c.Body()

		if json.Valid(body) {
			reqBody = json.RawMessage(body)
		} else {
			reqBody = string(body)
		}

		var respBody any
		body = c.Response().Body()

		if json.Valid(body) {
			respBody = json.RawMessage(body)
		} else {
			respBody = string(body)
		}

		payload["request"] = map[string]any{"method": c.Method(), "path": path, "host": c.Hostname(), "body": reqBody, "headers": c.Response().Header.String(), "queries": c.Queries()}
		payload["response"] = map[string]any{"body": respBody, "headers": c.Response().Header.String(), "status": statusCode}
		logData["payload"] = payload

		/*enc := json.NewEncoder(output)
		enc.SetEscapeHTML(false)
		err := enc.Encode(&logData)
		if err != nil {
			return -1, err
		}
		*/
		helpers.LogInfo(context.Background(), l, loc, logCtx, statusCode, msg, payload)
		output.Set([]byte{})
		return output.Len(), nil

	}
}

/*
func formatErrorLog(output logger.Buffer, _ *fiber.Ctx, data *logger.Data, extraParam string) (int, error) {
	if data.ChainErr != nil {
		return output.WriteString(" " + data.ChainErr.Error())
	}

	return 0, nil

}
*/
