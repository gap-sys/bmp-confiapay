package models

import "encoding/json"

type CachedCobranca struct {
	UrlWebhook    string `json:"urlWebhook"`
	CodLiquidacao string `json:"codLiquidacao"`
}

func (c CachedCobranca) MarshalBinary() ([]byte, error) {
	return json.Marshal(c)

}

func (c *CachedCobranca) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)

}
