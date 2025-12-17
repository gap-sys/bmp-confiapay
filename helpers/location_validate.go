package helpers

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// Sigla dos estados brasileiros
var estados = map[string]string{
	"AC": "Acre",
	"AL": "Alagoas",
	"AP": "Amapá",
	"AM": "Amazonas",
	"BA": "Bahia",
	"CE": "Ceará",
	"DF": "Distrito Federal",
	"ES": "Espírito Santo",
	"GO": "Goiás",
	"MA": "Maranhão",
	"MT": "Mato Grosso",
	"MS": "Mato Grosso do Sul",
	"MG": "Minas Gerais",
	"PA": "Pará",
	"PB": "Paraíba",
	"PR": "Paraná",
	"PE": "Pernambuco",
	"PI": "Piauí",
	"RJ": "Rio de Janeiro",
	"RN": "Rio Grande do Norte",
	"RS": "Rio Grande do Sul",
	"RO": "Rondônia",
	"RR": "Roraima",
	"SC": "Santa Catarina",
	"SP": "São Paulo",
	"SE": "Sergipe",
	"TO": "Tocantins",
}

// Verifica se a UF inserida é válida de acordo com "estados"
func ValidateUF(uf string) error {

	if _, ok := estados[uf]; !ok {

		return errors.New("estado invalido")
	}
	return nil
}

// Chama uma API externa para verificar se a cidade informada é válida. A URL da API será definida na varável de ambiente "CITIES_URL"
func ValidateCity(city string) error {

	if len(strings.Split(city, " ")) > 1 {
		city = strings.Replace(city, " ", "-", -1)
	}
	url := fmt.Sprintf("%s/%s", os.Getenv("CITIES_URL"), city)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("erro interno ao consultar api de cidades:%s", err.Error())
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("erro interno ao processar respostas da api de cidades:%s", err.Error())
	}
	if len(body) <= 2 {
		return errors.New("cidade invalida")
	}

	return nil
}
