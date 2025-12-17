package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// Envia "ping" para um endpoint e espera "pong" como resposta. Será utilizado para testar conexões.
func PingValidation(url string) error {
	var requestBody, response = map[string]string{"ping": "ping"}, make(map[string]string)
	body, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	var client = &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("requisicao retornou %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return err
	}

	pong := response["result"]
	if pong != "pong" {
		return fmt.Errorf("retorno incorreto do webhook")
	}

	return nil

}
