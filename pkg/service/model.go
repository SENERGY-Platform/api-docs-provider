package service

import (
	"encoding/json"
)

const (
	swaggerHostKey        = "host"
	swaggerBasePathKey    = "basePath"
	swaggerSchemesKey     = "schemes"
	swaggerPathsKey       = "paths"
	swaggerDefinitionsKey = "definitions"
)

var commonSwaggerKeys = map[string]struct{}{
	"swagger": {},
	"info":    {},
	"openapi": {},
	"paths":   {},
}

type docWrapper struct {
	basePath string
	doc      map[string]json.RawMessage
}
