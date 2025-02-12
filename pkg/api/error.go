package api

import (
	"errors"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/models"
	"net/http"
)

func GetStatusCode(err error) int {
	var nfe *models.NotFoundError
	if errors.As(err, &nfe) {
		return http.StatusNotFound
	}
	var iie *models.InvalidInputError
	if errors.As(err, &iie) {
		return http.StatusBadRequest
	}
	var ie *models.InternalError
	if errors.As(err, &ie) {
		return http.StatusInternalServerError
	}
	var rbe *models.ResourceBusyError
	if errors.As(err, &rbe) {
		return http.StatusConflict
	}
	return 0
}
