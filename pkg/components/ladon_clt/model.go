package ladon_clt

type accessRequest struct {
	Resource string `json:"resource"` // resource that access is requested to
	Action   string `json:"action"`   // action that is requested on the resource
	Subject  string `json:"subject"`  // subject that is requesting access
}

type accessResponse struct {
	Result bool   `json:"result"`
	Error  string `json:"error"`
}
