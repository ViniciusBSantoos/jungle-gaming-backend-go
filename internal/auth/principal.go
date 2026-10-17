package auth

type Principal struct {
	Subject  string
	Username string
	Roles    map[string]ClientRoles
}

func NewPrincipal(claims *JWTClaims) Principal {
	return Principal{
		Subject:  claims.Subject,
		Username: claims.PreferredUser,
		Roles:    claims.ResourceAccess,
	}
}

func (p Principal) HasRole(clientID, role string) bool {
	client, exists := p.Roles[clientID]
	if !exists {
		return false
	}

	for _, currentRole := range client.Roles {
		if currentRole == role {
			return true
		}
	}

	return false
}
