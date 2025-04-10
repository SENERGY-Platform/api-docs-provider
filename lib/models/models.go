package models

type SwaggerItem struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Version  string   `json:"version"`
	ExtPaths []string `json:"ext_paths"`
}

type AsyncapiItem struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Version string `json:"version"`
}
