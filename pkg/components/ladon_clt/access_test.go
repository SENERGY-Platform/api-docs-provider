package ladon_clt

import (
	"reflect"
	"testing"
)

func Test_newUserAccessRequest(t *testing.T) {
	a := []userAccessRequest{
		{
			Method:      "GET",
			Endpoint:    "/x/y",
			orgMethod:   "Get",
			orgEndpoint: "x/y",
		},
		{
			Method:      "PUT",
			Endpoint:    "/x/y",
			orgMethod:   "Put",
			orgEndpoint: "x/y",
		},
		{
			Method:      "GET",
			Endpoint:    "/y",
			orgMethod:   "GET",
			orgEndpoint: "/y",
		},
	}
	b := newUserAccessRequest(map[string][]string{
		"x/y": {"Get", "Put"},
		"/y":  {"GET"},
	})
	if !reflect.DeepEqual(a, b) {
		t.Errorf("expected: %v, got: %v", a, b)
	}
}

func Test_getUserAccessResult(t *testing.T) {
	a := map[string][]string{
		"x/y": {"Get"},
		"/y":  {"GET"},
	}
	req := []userAccessRequest{
		{
			Method:      "GET",
			Endpoint:    "/x/y",
			orgMethod:   "Get",
			orgEndpoint: "x/y",
		},
		{
			Method:      "PUT",
			Endpoint:    "/x/y",
			orgMethod:   "Put",
			orgEndpoint: "x/y",
		},
		{
			Method:      "GET",
			Endpoint:    "/y",
			orgMethod:   "GET",
			orgEndpoint: "/y",
		},
	}
	res := []bool{
		true,
		false,
		true,
	}
	b, err := getUserAccessResult(req, res)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Errorf("expected: %v, got: %v", a, b)
	}
}
