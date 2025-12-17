package helpers

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-playground/locales/en_US"
	"github.com/go-playground/locales/pt_BR"
	"github.com/go-playground/validator/v10"

	ut "github.com/go-playground/universal-translator"
)

// Traduções para as tags de erros retornados pelo validator.
var validationMessages = map[string]string{
	"required":                   "{0} é obrigatório.",
	"numeric":                    "{0} deve ser numérico.",
	"min":                        "{0} deve ter no mínimo {1} {2}.",
	"max":                        "{0} deve ter no máximo {1} {2}.",
	"len":                        "{0} deve ter exatamente {1} {2}.",
	"uppercase":                  "{0} deve estar em maiúsculas.",
	"alpha":                      "{0} deve conter apenas letras.",
	"alpha_num":                  "{0} deve conter apenas letras e números.",
	"email":                      "{0} deve ser um e-mail válido.",
	"url":                        "{0} deve ser uma URL válida.",
	"uuid":                       "{0} deve ser um UUID válido.",
	"eqfield":                    "{0} e {1} devem ser iguais.",
	"gt":                         "{0} deve ser maior que {1}.",
	"gte":                        "{0} deve ser maior ou igual a {1}.",
	"lt":                         "{0} deve ser menor que {1}.",
	"lte":                        "{0} deve ser menor ou igual a {1}.",
	"unique":                     "{0} já existe.",
	"ip":                         "{0} deve ser um endereço IP válido.",
	"cidr":                       "{0} deve ser um CIDR válido.",
	"ipv4":                       "{0} deve ser um endereço IPv4 válido.",
	"ipv6":                       "{0} deve ser um endereço IPv6 válido.",
	"gtfield":                    "{0} deve ser maior que {1}.",
	"gtefield":                   "{0} deve ser maior ou igual a {1}.",
	"ltfield":                    "{0} deve ser menor que {1}.",
	"ltefield":                   "{0} deve ser menor ou igual a {1}.",
	"alpha_spaces":               "{0} deve conter apenas letras e espaços.",
	"alpha_dash":                 "{0} deve conter apenas letras, números, traços e underscores.",
	"alphanumeric":               "{0} deve conter apenas letras e números.",
	"hex":                        "{0} deve ser um valor hexadecimal.",
	"hexadecimal":                "{0} deve ser um valor hexadecimal.",
	"json":                       "{0} deve ser uma string JSON válida.",
	"array":                      "{0} deve ser uma lista.",
	"date":                       "{0} deve ser uma data válida.",
	"timezone":                   "{0} deve ser um fuso horário válido.",
	"file":                       "{0} deve ser um arquivo válido.",
	"image":                      "{0} deve ser uma imagem válida.",
	"mime":                       "{0} deve ser um tipo de arquivo válido.",
	"validate_city":              "{0} deve ser uma cidade válida",
	"validate_uf":                "{0} deve ser um estado válido",
	"validate_cpf":               "{0} deve ser um cpf válido",
	"validate_date":              "{0} deve ser uma data no formato {1}",
	"datetime":                   "{0} deve ser uma data no formato {1}",
	"date_before":                "{0} deve ser uma data antes de {1}",
	"date_after":                 "{0} deve ser uma data depois de {1}",
	"required_without":           "{0} é requerido caso {1} seja vazio",
	"required_with":              "{0} é requerido caso {1} não seja vazio",
	"required_if":                "{0} é requerido caso {1}",
	"lt_today":                   "{0} não pode ser maior que a data atual",
	"date_gap":                   "{0} não deve ser {1}",
	"oneof":                      "{0} deve ser um dos seguintes valores:[{1}]",
	"ping":                       "{0} deve ser um endereço válido",
	"excluded_unless":            "{0} não deve ser passado caso {1} não seja vazio",
	"validate_documento_federal": "{0} deve ser um documento válido",
}

// Unidades para personalizar a mensagem final das validações de máx,min e len
var unities = map[reflect.Kind]string{
	reflect.Slice:                      "elementos",
	reflect.String:                     "caracteres",
	reflect.TypeOf(time.Time{}).Kind(): "dias",
}

// Cria um tradutr PT_BR.
func NewTranslator() (ut.Translator, error) {
	uni := ut.New(en_US.New(), pt_BR.New())
	translator, found := uni.GetTranslator("pt_BR")
	if !found {
		return nil, errors.New("erro ao inicializar tradutor")
	}
	return translator, nil

}

// Associa o tradutor ao validator
func ConfigTranslator(v *validator.Validate, t ut.Translator, strct any) error {
	for tag, message := range validationMessages {
		err := v.RegisterTranslation(tag, t,
			func(ut ut.Translator) error {
				return ut.Add(tag, message, true)
			},
			func(ut ut.Translator, fe validator.FieldError) string {
				var param string
				switch fe.Tag() {
				case "date_after", "date_before", "date_gap":
					param = formatDateFields(fe)
				case "required_without", "required_with":
					param = fieldToCamelCase(fe.Param())
				case "required_if":
					params := strings.Split(fe.Param(), " ")
					params[0] = fieldToCamelCase(params[0])
					param = fmt.Sprintf("%s=%s", params[0], params[1])

				case "excluded_unless":
					params := strings.Split(fe.Param(), " ")
					param = params[0]

				default:
					param = fe.Param()

				}

				jsonTag := ExtractStructTag(fe.Field(), "json", strct, true)
				tags := strings.Split(jsonTag, ",")
				if len(tags) > 1 {
					jsonTag = tags[0]

				}
				unity := findUnity(fe)
				return replacePlaceholders(message, jsonTag, param, unity)
			})
		if err != nil {
			return err
		}
	}

	return nil
}

// Formata as traduções adicionando nome do campo e possíveis parâmetros
func replacePlaceholders(message, field, value, unity string) string {
	// Substitui o {0} pelo nome do campo
	message = strings.Replace(message, "{0}", field, -1)

	// Substitui o {1} pelo valor do parâmetro, se necessário

	if value != "" {
		message = strings.Replace(message, "{1}", value, -1)
	}
	if unity != "" {
		message = strings.Replace(message, "{2}", unity, -1)

	}

	return message
}

// Converte diferentes unidades e personaliza a mensagem das tags max e min
func findUnity(fe validator.FieldError) string {
	unity := ""
	tag := fe.Tag()
	kind := fe.Kind()
	switch tag {
	case "min", "max", "len":
		unity := unities[kind]
		return unity
	default:
		return unity
	}

}

// Formata a mensagem para o validator de gap
func formatDateFields(fe validator.FieldError) string {
	param := fe.Param()
	params := strings.Split(param, ":")
	if len(params) == 3 {
		var direction = map[string]string{
			"after":  "maior",
			"before": "menor",
		}

		return fmt.Sprintf("%s que %s dias comparado a %s", direction[params[2]], params[1], fieldToCamelCase(params[0]))
	}

	fieldName := fieldToCamelCase(fe.Param())
	return fieldName
}

// Converte a primeira letra dos nomes de campos de struct(que se espera estarem em maiúsculo)para minúsculo, seguindo assim o padrão camelCase.
func fieldToCamelCase(field string) string {
	if isAllUpperCase(field) { // Caso o campo esteja todo em maiúsculo(uma sigla por exemplo), será todo convertido em minúsculo.
		return strings.ToLower(field)
	}
	var camelCase string
	camelCase += string(strings.ToLower(field)[0])
	camelCase += field[1:]

	return camelCase
}

// Checa se uma string está toda em maiúsculo
func isAllUpperCase(s string) bool {
	for _, char := range s {
		if char < 0x41 || char > 0x5A { // Code points em UTF-8 de A a Z maiúsculo. Se não estiver neste intervalo não é uma letra maiúscula.
			return false
		}

	}
	return true
}
