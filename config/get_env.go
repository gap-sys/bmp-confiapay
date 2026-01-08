package config

import (
	"log"
	"os"
)

func getEnv(name string) string {
	envVar := os.Getenv(name)

	if envVar == "" {
		log.Fatalf("Valor para variável %s não encontrada no .env", name)
	}
	return envVar
}
