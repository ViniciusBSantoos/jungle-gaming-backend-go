package auth

import "testing"

func TestParseAccessClaims(t *testing.T) {
	data := []byte(`{
        "sub": "jogador-123",
        "preferred_username": "jungle-tester",
        "resource_access": {
            "jungle-backend-api": {
                "roles": ["wallet:read", "wallet:write"]
            }
        }
    }`)

	claims, err := ParseAccessClaims(data)
	if err != nil {
		t.Fatalf("Erro ao interpretar claims: %v", err)
	}

	if claims.Subject != "jogador-123" {
		t.Error("Subject incorreto")
	}

	if !claims.HasRole("jungle-backend-api", "wallet:read") {
		t.Error("Permissao wallet:read nao encontrada")
	}

	if !claims.HasRole("jungle-backend-api", "wallet:write") {
		t.Error("Permissao wallet:write nao encontrada")
	}
}

func TestHasRoleMissing(t *testing.T) {
	claims := AccessClaims{
		Subject: "jogador-123",
		ResourceAccess: map[string]ClientRoles{
			"jungle-backend-api": {
				Roles: []string{"wallet:read"},
			},
		},
	}

	if claims.HasRole("jungle-backend-api", "wallet:write") {
		t.Error("Nao deveria autorizar wallet:write")
	}

	if claims.HasRole("outro-cliente", "wallet:read") {
		t.Error("Nao deveria autorizar outro cliente")
	}
}

func TestParseAccessClaimsMissingSubject(t *testing.T) {
	data := []byte(`{
        "resource_access": {
            "jungle-backend-api": {
                "roles": ["wallet:read"]
            }
        }
    }`)

	_, err := ParseAccessClaims(data)

	if err == nil {
		t.Error("Deveria rejeitar subject vazio")
	}
}

func TestValidateTokenTypeBearer(t *testing.T) {
	claims := AccessClaims{
		TokenType: "Bearer",
	}

	if err := claims.ValidateTokenType(); err != nil {
		t.Errorf("Token Bearer deveria ser aceito: %v", err)
	}
}

func TestValidateTokenTypeID(t *testing.T) {
	claims := AccessClaims{
		TokenType: "ID",
	}

	if err := claims.ValidateTokenType(); err == nil {
		t.Error("Token ID deveria ser rejeitado")
	}
}

func TestValidateTokenTypeMissing(t *testing.T) {
	claims := AccessClaims{}

	if err := claims.ValidateTokenType(); err == nil {
		t.Error("Token sem tipo deveria ser rejeitado")
	}
}
