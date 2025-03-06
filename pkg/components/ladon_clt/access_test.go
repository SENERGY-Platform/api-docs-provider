/*
 * Copyright 2025 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
