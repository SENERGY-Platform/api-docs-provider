package service

import (
	"encoding/json"
)

const (
	swaggerKey            = "swagger"
	swaggerInfoKey        = "info"
	swaggerOpenApiKey     = "openapi"
	swaggerHostKey        = "host"
	swaggerBasePathKey    = "basePath"
	swaggerSchemesKey     = "schemes"
	swaggerPathsKey       = "paths"
	swaggerDefinitionsKey = "definitions"
)

var swaggerV2Keys = []string{
	swaggerKey,
	swaggerInfoKey,
	swaggerPathsKey,
}

var swaggerV3Keys = []string{
	swaggerInfoKey,
	swaggerOpenApiKey,
	swaggerPathsKey,
}

type docWrapper struct {
	basePath string
	doc      map[string]json.RawMessage
}
