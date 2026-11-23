package response

// AuthTokenResponse is the token payload returned on successful login or token refresh.
type AuthTokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// AuthLoginResponse wraps the actor profile with tokens.
type AuthLoginResponse[T any] struct {
	Actor        T                 `json:"actor"`
	AccessToken  string            `json:"accessToken"`
	RefreshToken string            `json:"refreshToken"`
}
