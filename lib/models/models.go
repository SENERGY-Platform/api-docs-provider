package models

const (
	HeaderRequestID = "X-Request-ID"
	HeaderApiVer    = "X-Api-Version"
	HeaderSrvName   = "X-Service-Name"
)

type SwaggerItem struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Version     string   `json:"version"`
	ExtPaths    []string `json:"ext_paths"`
	Description string   `json:"description"`
}

type AsyncapiItem struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Version     string `json:"version"`
	Description string `json:"description"`
}
