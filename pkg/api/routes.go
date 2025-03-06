package api

import (
	gin_mw "github.com/SENERGY-Platform/gin-middleware"
)

var routes = gin_mw.Routes[Service]{
	getSwaggerDocsH,
	patchStorageRefreshH,
	getStorageListH,
	getInfoH,
	getHealthCheckH,
	getSwaggerDocH,
}
