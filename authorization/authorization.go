package authorization

// Token represents SSO token from Chat API's perspective.
type Token struct {
	// AccessToken is a customer access token returned by LiveChat OAuth Server.
	AccessToken string
	// Region is a datacenter for LicenseID (`dal` or `fra`).
	Region string
	// Type specifies whether it is Bearer or Basic token type.
	Type TokenType
	// OrganizationID specifies ID of organization which owns the token.
	OrganizationID string
}

// TokenGetter is called by each API method to obtain valid Token.
// If TokenGetter returns nil, the method won't be executed on API.
type TokenGetter func() *Token
