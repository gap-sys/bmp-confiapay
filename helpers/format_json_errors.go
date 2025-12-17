package helpers

import (
	"fmt"
	"regexp"
	"strings"
)

func ParseJsonError(errMsg string) string {
	if errMsg == "unexpected end of JSON input" {
		return "JSON inválido:Objeto JSON não foi fechado corretamente."
	}
	// Caso 1: cannot unmarshal ... into Go struct field ...
	reField := regexp.MustCompile(`struct field [\w\.]+\.([\w]+) of type ([\[\]\w{}: ]+)`)
	if matches := reField.FindStringSubmatch(errMsg); len(matches) == 3 {
		campo := matches[1]
		tipo := strings.TrimSpace(matches[2])
		return fmt.Sprintf("%s deve ser do tipo %s", campo, tipo)
	}

	// Caso 2: invalid character ...
	reInvalidChar := regexp.MustCompile(`invalid character '(.+)'`)
	if matches := reInvalidChar.FindStringSubmatch(errMsg); len(matches) == 2 {
		return fmt.Sprintf("JSON inválido: caractere inesperado '%s'", matches[1])
	}

	// Caso 3: cannot unmarshal ... into Go value of type ...
	reRoot := regexp.MustCompile(`cannot unmarshal \w+ into Go value of type ([\w\.]+)`)
	if matches := reRoot.FindStringSubmatch(errMsg); len(matches) == 2 {
		return fmt.Sprintf("JSON inválido: esperado tipo %s", matches[1])
	}

	// Caso não reconhecido -> devolve mensagem genérica
	return "JSON inválido"
}
