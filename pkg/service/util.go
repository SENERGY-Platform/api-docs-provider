package service

const (
	swaggerHostKey     = "host"
	swaggerBasePathKey = "basePath"
	swaggerSchemesKey  = "schemes"
	swaggerPathsKey    = "paths"
)

func stringInSlice(a string, sl []string) bool {
	for _, b := range sl {
		if b == a {
			return true
		}
	}
	return false
}
