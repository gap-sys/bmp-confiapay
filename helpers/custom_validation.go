package helpers

import (
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

var (
	webhookCache = make(map[string]bool)
)

var customValidation = map[string]validator.Func{

	"validate_city": func(fl validator.FieldLevel) bool {
		city := fl.Field().String()
		return ValidateCity(city) == nil

	},

	"validate_cpf": func(fl validator.FieldLevel) bool {
		cpf := fl.Field().String()
		return ValidateCPF(cpf) == nil
	},
	"validate_cnpj": func(fl validator.FieldLevel) bool {
		cnpj := fl.Field().String()
		return ValidateCNPJ(cnpj) == nil
	},

	"validate_documento_federal": func(fl validator.FieldLevel) bool {

		doc := strings.Split(fl.Field().String(), "#")
		if len(doc) != 2 {
			return false
		}

		switch doc[0] {
		case "CPF":
			return ValidateCPF(doc[1]) == nil
		case "CNPJ":
			return ValidateCNPJ(doc[1]) == nil
		default:
			return false
		}

	},
	"validate_uf": func(fl validator.FieldLevel) bool {
		uf := fl.Field().String()
		return ValidateUF(uf) == nil
	},

	"date_before": func(fl validator.FieldLevel) bool {
		date := fl.Field().String()
		before := fl.Parent().FieldByName(fl.Param()).String()
		format := ExtractStructTag(fl.StructFieldName(), "validate", fl.Parent().Interface(), false)
		_, format, _ = strings.Cut(format, "datetime=")
		format, _, _ = strings.Cut(format, ",")

		return DateBefore(date, before, format)

	},
	"date_after": func(fl validator.FieldLevel) bool {
		date := fl.Field().String()
		after := fl.Parent().FieldByName(fl.Param()).String()
		format := ExtractStructTag(fl.StructFieldName(), "validate", fl.Parent().Interface(), false)
		_, format, _ = strings.Cut(format, "datetime=")
		format, _, _ = strings.Cut(format, ",")
		return DateAfter(date, after, format)

	},
	"date_gap": func(fl validator.FieldLevel) bool {
		date := fl.Field().String()
		params := strings.Split(fl.Param(), ":")
		reference := fl.Parent().FieldByName(params[0]).String()
		gap, _ := strconv.ParseInt(params[1], 10, 64)
		orientation := params[2]
		format := ExtractStructTag(fl.StructFieldName(), "validate", fl.Parent().Interface(), false)
		_, format, _ = strings.Cut(format, "datetime=")
		format, _, _ = strings.Cut(format, ",")
		return DateGap(date, reference, format, orientation, gap)

	},
	"lt_today": func(fl validator.FieldLevel) bool {

		d := fl.Field().String()
		format := ExtractStructTag(fl.StructFieldName(), "validate", fl.Parent().Interface(), false)
		_, format, _ = strings.Cut(format, "datetime=")
		format, _, _ = strings.Cut(format, ",")
		date, _ := time.Parse(format, d)
		return !date.After(time.Now())

	},
	"ping": func(fl validator.FieldLevel) bool {
		url := fl.Field().String()
		cached := webhookCache[url]
		if cached {
			return true
		}

		err := PingValidation(url)
		if err != nil {
			return false
		}

		webhookCache[url] = true
		return true

	},
	"validate_cargo": func(fl validator.FieldLevel) bool {
		cargo := fl.Field().Int()
		_, ok := cargos[int(cargo)]
		return ok
	},
}

// Registra validações customizadas que não estão disponíveis no package validator
func registerCustomValidation(v *validator.Validate, tags ...string) {

	for _, tag := range tags {
		validation, ok := customValidation[tag]
		if !ok {
			panic("validação não encontrada")
		}
		v.RegisterValidation(tag, validation)
	}
}
