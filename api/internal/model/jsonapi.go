package model

// JSONAPIUserSignupResponse represents the user for this application
//
// Refer to JSON:API specification at https://jsonapi.org/
//
// swagger:model JSONAPIUserSignupResponse
type JSONAPIUserSignupResponse struct {
	// json:api data property
	//
	// required: true
	// read-only: true
	Data struct {
		// json:api attributes property
		//
		// required: true
		// read-only: true
		Attributes User `json:"attributes"`
		// json:api id property
		//
		// required: true
		// read-only: true
		// example: a2aee5e6-05a0-438c-9276-4ba406b7bf9e
		ID string `json:"id"`
		// json:api links property
		//
		// required: true
		// read-only: true
		Links struct {
			// json:api self property
			//
			// required: true
			// read-only: true
			// example: /users/a2aee5e6-05a0-438c-9276-4ba406b7bf9e
			Self string `json:"self"`
		} `json:"links"`
		// json:api type property
		//
		// required: true
		// read-only: true
		// example: users
		Type string `json:"type"`
	} `json:"data"`
	// json:api jsonapi property
	//
	// required: true
	// read-only: true
	Jsonapi struct {
		// json:api version property
		//
		// required: true
		// read-only: true
		// example: 1.0
		Version string `json:"version"`
	} `json:"jsonapi"`
	// json:api links property
	//
	// required: true
	// read-only: true
	Links struct {
		// json:api links self property
		//
		// required: true
		// read-only: true
		// example: /users/a2aee5e6-05a0-438c-9276-4ba406b7bf9e
		Self string `json:"self"`
	} `json:"links"`
}

// JSONAPIUserSignup represents the user signup for this application
//
// Refer to JSON:API specification at https://jsonapi.org/
//
// swagger:model JSONAPIUserSignup
type JSONAPIUserSignup struct {
	// json:api data property
	//
	// required: true
	Data struct {
		// json:api attributes property
		//
		// required: true
		Attributes User `json:"attributes"`
		// json:api type property
		//
		// required: true
		// example: users
		Type string `json:"type"`
	} `json:"data"`
	// json:api jsonapi property
	//
	// required: true
	Jsonapi struct {
		// json:api version property
		//
		// required: true
		// example: 1.0
		Version string `json:"version"`
	} `json:"jsonapi"`
}
