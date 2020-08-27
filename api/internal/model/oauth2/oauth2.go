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
	ID     string `json:"client_id"`
	Name   string `json:"name"`
	Secret string `json:"-"`
}

type ResourceOwnerGrantedPermissions struct {
	ID       string   `json:"user_id"`
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
