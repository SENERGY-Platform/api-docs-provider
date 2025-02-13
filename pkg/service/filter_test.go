package service

import (
	"encoding/json"
	"reflect"
	"testing"
)

func Test_transformDoc(t *testing.T) {
	orgDoc := []byte("{\"host\": \"org\", \"basePath\": \"org\", \"schemes\": [\"http\"]}")
	srv := New(nil, nil, nil, nil, nil, 0, "test", "")
	aRaw := []byte("{\"host\": \"test\", \"basePath\": \"test\", \"schemes\": [\"http\"]}")
	var a map[string]json.RawMessage
	if err := json.Unmarshal(aRaw, &a); err != nil {
		t.Fatal(err)
	}
	b, err := srv.transformDoc(orgDoc, "test")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(b, a) {
		t.Errorf("got %v, expected %v", b, a)
	}
	t.Run("missing schemes", func(t *testing.T) {
		orgDoc := []byte("{\"host\": \"org\", \"basePath\": \"org\"}")
		aRaw := []byte("{\"host\": \"test\", \"basePath\": \"test\", \"schemes\": [\"https\"]}")
		var a map[string]json.RawMessage
		if err := json.Unmarshal(aRaw, &a); err != nil {
			t.Fatal(err)
		}
		b, err := srv.transformDoc(orgDoc, "test")
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(b, a) {
			t.Errorf("got %v, expected %v", b, a)
		}
	})
}
