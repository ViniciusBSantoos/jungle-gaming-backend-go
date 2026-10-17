package auth

import "testing"

func TestNewConfigValid(t *testing.T) {
	config, err := NewConfig(
		"http://localhost:8081/realms/jungle-gaming",
		"jungle-backend-api",
	)

	if err != nil {
		t.Fatalf("Configuracao valida rejeitada: %v", err)
	}

	if config.Audience != "jungle-backend-api" {
		t.Error("Audience incorreta")
	}
}

func TestNewConfigMissingIssuer(t *testing.T) {
	_, err := NewConfig("", "jungle-backend-api")

	if err == nil {
		t.Error("Deveria rejeitar issuer vazio")
	}
}

func TestNewConfigMissingAudience(t *testing.T) {
	_, err := NewConfig(
		"http://localhost:8081/realms/jungle-gaming",
		"",
	)

	if err == nil {
		t.Error("Deveria rejeitar audience vazia")
	}
}
