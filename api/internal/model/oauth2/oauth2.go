package oauth2

// swagger:model AuthorizationRequest
type AuthorizationRequest struct {
	// required: true
	// example: 352b6e64-e498-4307-b64d-ec9e5b9da65c
	ClientID string `json:"client_id"`
	// required: true
	// example: http://accounts.example.com/oauth2-callback
	RedirectURI string `json:"redirect_uri"`
	// required: true
	// example: read:profile,write:profile
	Scope string `json:"scope"`
	// required: true
	// example: code
	ResponseType string `json:"response_type"`
	// required: true
	// example: query
	ResponseMode string `json:"response_mode"`
	// example: sto7zydoa6o
	Nonce string `json:"nonce"`
	State string `json:"state"`
}

// swagger:model AuthorizationResponse
type AuthorizationResponse struct {
	// example:5c001023-a3f4-4c68-b39b-07f040bbeed4
	Code string `json:"code"`
	// example: http://accounts.example.com/oauth2-callback
	RedirectURI string `json:"redirect_uri"`
	// example: sto7zydoa6o
	Nonce string `json:"nonce"`
	State string `json:"state"`
}

// Access Token Response
//
// Foo
//
// swagger:model AccessTokenResponse
type AccessTokenResponse struct {
	// required: true
	// example: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiI3b25ldGVsbGEiLCJleHAiOjE2MDQ4MjQwODEsImlhdCI6MTYwNDgyMDQ4MSwiaXNzIjoiN29uZXRlbGxhIiwic3ViIjoiNjNjMzJkMTAtYzA2ZC00NmQ4LWI5ZTUtNGU0ZDhlZDk3MzJlIn0.NFnO9s_hujiqDHFbH7RvLoIeseIzGs0VU05whMq0x7U
	AccessToken string `json:"access_token"`
	// required: true
	// example: bearer
	TokenType string `json:"token_type"`
	// example: 3600
	// required: true
	ExpiresIn string `json:"expires_in"`
	// example: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiI3b25ldGVsbGEiLCJleHAiOjE2MDQ4MjQwODEsImlhdCI6MTYwNDgyMDQ4MSwiaXNzIjoiN29uZXRlbGxhIiwic3ViIjoiNjNjMzJkMTAtYzA2ZC00NmQ4LWI5ZTUtNGU0ZDhlZDk3MzJlIn0.NFnO9s_hujiqDHFbH7RvLoIeseIzGs0VU05whMq0x7U
	RefreshToken string `json:"refresh_token"`
	// required: true
	// example: read:profile,write:profile
	Scope string `json:"scope"`
}

type Client struct {
	ID     string `db:"client_id" json:"client_id"`
	Name   string `db:"name"      json:"name"`
	Secret string `db:"secret"    json:"-"`
}

type ResourceOwnerGrantedPermissions struct {
	UserID   string   `json:"user_id"`
	ClientID string   `json:"client_id"`
	Scope    []string `json:"scope"`
}

type Resource struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Permissions []struct {
		Name string `json:"name"`
		Desc string `json:"desc"`
	} `json:"permissions"`
}

type UserGrants struct {
	UserID   string `db:"user_id"    json:"user_id"`
	ClientID string `db:"client_id"  json:"client_id"`
	Scope    string `db:"scope"      json:"scope"`
}

type AuthorizationCode struct {
	Code      string `db:"code"`
	ClientID  string `db:"client_id"`
	CreatedAt int64  `db:"created_at"`
	UserID    string `db:"user_id"`
}

type AccessToken struct {
	TokenID string `db:"token_id"`
	UserID  string `db:"user_id"`
	Token   string `db:"token"`
}
