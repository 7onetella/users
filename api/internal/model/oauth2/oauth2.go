package oauth2

type AuthorizationRequest struct {
	ClientID     string `json:"client_id"`
	RedirectURI  string `json:"redirect_uri"`
	Scope        string `json:"scope"`
	ResponseType string `json:"response_type"`
	ResponseMode string `json:"response_mode"`
	Nonce        string `json:"nonce"`
	State        string `json:"state"`
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
