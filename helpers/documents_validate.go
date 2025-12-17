package helpers

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// Valida o CPF passado de acordo com padrão de caracteres e validação do dígito verificador
func ValidateCPF(cpf string) error {
	cpf = regexp.MustCompile("[^0-9]").ReplaceAllString(cpf, "")
	if len(cpf) != 11 || cpf == "00000000000" {
		return errors.New("CPF invalido")
	}
	for i := 9; i < 11; i++ {
		sum := 0
		for j := 0; j < i; j++ {
			num, _ := strconv.Atoi(string(cpf[j]))
			sum += num * (i + 1 - j)
		}

		checkDigit := (sum * 10) % 11
		if checkDigit == 10 {
			checkDigit = 0
		}
		if checkDigit != int(cpf[i]-'0') {
			return errors.New("CPF invalido")
		}
	}
	return nil
}

// Valida o CNPJ passado de acordo com padrão de caracteres e validação do dígito verificador
func ValidateCNPJ(cnpj string) error {
	cnpj = regexp.MustCompile("[^0-9]").ReplaceAllString(cnpj, "")
	if len(cnpj) != 14 || cnpj == "00000000000000" {
		return errors.New("CNPJ invalido")
	}
	var weightFirst = []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	var weightSecond = []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	for i := 12; i < 14; i++ {
		sum := 0
		for j := 0; j < i; j++ {
			num, _ := strconv.Atoi(string(cnpj[j]))
			if i == 12 {
				sum += num * weightFirst[j]
			} else {
				sum += num * weightSecond[j]
			}
		}
		checkDigit := sum % 11
		if checkDigit < 2 {
			checkDigit = 0
		} else {
			checkDigit = 11 - checkDigit
		}
		if checkDigit != int(cnpj[i]-'0') {
			return errors.New("CNPJ invalido")
		}
	}
	return nil
}

// Verifica se uma string é composta apenas de caracteres numéricos
func IsNumeric(str string) bool {
	for _, char := range str {
		if !unicode.IsDigit(char) {
			return false
		}
	}
	return true
}

// Remove caracteres não numéricos de uma string
func RemoveNonNumeric(str string) string {
	var newStr strings.Builder
	for _, char := range str {
		if unicode.IsDigit(char) {
			newStr.WriteRune(char)
		}
	}
	return newStr.String()

}
