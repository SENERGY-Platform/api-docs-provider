package ladon_clt

type roleAccessRequest struct {
	Resource string `json:"resource"` // resource that access is requested to
	Action   string `json:"action"`   // action that is requested on the resource
	Subject  string `json:"subject"`  // subject that is requesting access
}

type roleAccessResponse struct {
	Result bool   `json:"result"`
	Error  string `json:"error"`
}

type userAccessRequest struct {
	Method      string `json:"method"`
	Endpoint    string `json:"endpoint"`
	orgMethod   string
	orgEndpoint string
}

type userAccessResponse struct {
	Allowed []bool `json:"allowed"`
}
