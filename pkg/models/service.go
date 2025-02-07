package models

type Service struct {
	ID       string
	Host     string
	Port     int
	Protocol string
	ExtPaths []string
}
