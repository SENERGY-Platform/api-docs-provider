package service

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestService_filterDoc(t *testing.T) {
	f, err := os.Open("test/swagger.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var doc map[string]json.RawMessage
	if err = json.NewDecoder(f).Decode(&doc); err != nil {
		t.Fatal(err)
	}
	ladonClt := &ladonCltMock{}
	srv := New(nil, nil, nil, nil, ladonClt, 0, "", "")
	t.Run("include", func(t *testing.T) {
		ladonClt.TokenPolicies = map[string][]string{
			"/a": {"get"},
		}
		ok, err := srv.filterDoc(context.Background(), doc, "test", nil, "")
		if err != nil {
			t.Error(err)
		}
		if !ok {
			t.Error("expected true")
		}
	})
	t.Run("exclude", func(t *testing.T) {
		ladonClt.TokenPolicies = map[string][]string{}
		ok, err := srv.filterDoc(context.Background(), doc, "test", nil, "")
		if err != nil {
			t.Error(err)
		}
		if ok {
			t.Error("expected false")
		}
	})
}

func Test_getNewDefinitions(t *testing.T) {
	f, err := os.Open("test/swagger.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var doc map[string]json.RawMessage
	if err = json.NewDecoder(f).Decode(&doc); err != nil {
		t.Fatal(err)
	}
	oldDefs, err := getDocDefs(doc)
	if err != nil {
		t.Fatal(err)
	}
	t.Run("full", func(t *testing.T) {
		newDefs := getNewDefinitions(oldDefs, map[string]struct{}{"A": {}, "B": {}, "C": {}})
		if len(newDefs) != len(oldDefs) {
			t.Errorf("expected %d definitions, got %d", len(oldDefs), len(newDefs))
		}
		for key := range oldDefs {
			if _, ok := newDefs[key]; !ok {
				t.Errorf("missing definition '%s'", key)
			}
		}
	})
	t.Run("partial", func(t *testing.T) {
		newDefs := getNewDefinitions(oldDefs, map[string]struct{}{"A": {}})
		if len(newDefs) != 1 {
			t.Errorf("expected 1 definition, got %d", len(newDefs))
		}
		if _, ok := newDefs["A"]; !ok {
			t.Error("expected definition 'A'")
		}
	})
	t.Run("none", func(t *testing.T) {
		newDefs := getNewDefinitions(oldDefs, map[string]struct{}{})
		if len(newDefs) != 0 {
			t.Errorf("expected 0 definitions, got %d", len(newDefs))
		}
	})
}

func TestService_getNewPathsByRoles(t *testing.T) {
	ladonClt := &ladonCltMock{}
	srv := New(nil, nil, nil, nil, ladonClt, 0, "", "")
	f, err := os.Open("test/swagger.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var doc map[string]json.RawMessage
	if err = json.NewDecoder(f).Decode(&doc); err != nil {
		t.Fatal(err)
	}
	oldPaths, err := getDocPaths(doc)
	if err != nil {
		t.Fatal(err)
	}
	t.Run("full", func(t *testing.T) {
		ladonClt.RolePolicies = map[string]map[string]struct{}{
			"/a": {"get": struct{}{}, "post": struct{}{}},
			"/b": {"get": struct{}{}},
		}
		newPaths, allowedRefs, err := srv.getNewPathsByRoles(context.Background(), oldPaths, "", []string{"test"})
		if err != nil {
			t.Error(err)
		}
		if len(newPaths) != 2 {
			t.Errorf("expected 2 paths, got %d", len(newPaths))
		}
		for p, methods := range map[string][]string{
			"/a": {"get", "post"},
			"/b": {"get"},
		} {
			methods2, ok := newPaths[p]
			if !ok {
				t.Errorf("missing path '%s'", p)
			}
			if len(methods2) != len(methods) {
				t.Errorf("expected %d methods, got %d", len(methods), len(methods2))
			}
			for _, method := range methods {
				if _, ok := methods2[method]; !ok {
					t.Errorf("missing method '%s'", method)
				}
			}
		}
		if len(allowedRefs) != 3 {
			t.Errorf("expected 3 references, got %d", len(newPaths))
		}
		for _, s := range []string{"A", "B", "C"} {
			if _, ok := allowedRefs[s]; !ok {
				t.Errorf("missing reference '%s'", s)
			}
		}
	})
	t.Run("partial", func(t *testing.T) {
		ladonClt.RolePolicies = map[string]map[string]struct{}{
			"/a": {"get": struct{}{}},
		}
		newPaths, allowedRefs, err := srv.getNewPathsByRoles(context.Background(), oldPaths, "", []string{"test"})
		if err != nil {
			t.Error(err)
		}
		if len(newPaths) != 1 {
			t.Errorf("expected 1 path, got %d", len(newPaths))
		}
		methods, ok := newPaths["/a"]
		if !ok {
			t.Error("missing path '/a'")
		}
		if len(methods) != 1 {
			t.Errorf("expected 1 method, got %d", len(methods))
		}
		if _, ok = methods["get"]; !ok {
			t.Error("missing method 'get'")
		}
		if len(allowedRefs) != 1 {
			t.Errorf("expected 1 reference, got %d", len(newPaths))
		}
		if _, ok := allowedRefs["A"]; !ok {
			t.Error("missing reference 'A'")
		}
	})
	t.Run("none", func(t *testing.T) {
		ladonClt.RolePolicies = map[string]map[string]struct{}{}
		newPaths, allowedRefs, err := srv.getNewPathsByRoles(context.Background(), oldPaths, "", []string{"test"})
		if err != nil {
			t.Error(err)
		}
		if len(newPaths) != 0 {
			t.Errorf("expected 0 paths, got %d", len(newPaths))
		}
		if len(allowedRefs) != 0 {
			t.Errorf("expected 0 references, got %d", len(newPaths))
		}
	})
}

func TestService_getNewPathsByToken(t *testing.T) {
	ladonClt := &ladonCltMock{}
	srv := New(nil, nil, nil, nil, ladonClt, 0, "", "")
	f, err := os.Open("test/swagger.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var doc map[string]json.RawMessage
	if err = json.NewDecoder(f).Decode(&doc); err != nil {
		t.Fatal(err)
	}
	oldPaths, err := getDocPaths(doc)
	if err != nil {
		t.Fatal(err)
	}
	t.Run("full", func(t *testing.T) {
		ladonClt.TokenPolicies = map[string][]string{
			"/a": {"get", "post"},
			"/b": {"get"},
		}
		newPaths, allowedRefs, err := srv.getNewPathsByToken(context.Background(), oldPaths, "", "")
		if err != nil {
			t.Error(err)
		}
		if len(newPaths) != 2 {
			t.Errorf("expected 2 paths, got %d", len(newPaths))
		}
		for p, methods := range map[string][]string{
			"/a": {"get", "post"},
			"/b": {"get"},
		} {
			methods2, ok := newPaths[p]
			if !ok {
				t.Errorf("missing path '%s'", p)
			}
			if len(methods2) != len(methods) {
				t.Errorf("expected %d methods, got %d", len(methods), len(methods2))
			}
			for _, method := range methods {
				if _, ok := methods2[method]; !ok {
					t.Errorf("missing method '%s'", method)
				}
			}
		}
		if len(allowedRefs) != 3 {
			t.Errorf("expected 3 references, got %d", len(newPaths))
		}
		for _, s := range []string{"A", "B", "C"} {
			if _, ok := allowedRefs[s]; !ok {
				t.Errorf("missing reference '%s'", s)
			}
		}
	})
	t.Run("partial", func(t *testing.T) {
		ladonClt.TokenPolicies = map[string][]string{
			"/a": {"get"},
		}
		newPaths, allowedRefs, err := srv.getNewPathsByToken(context.Background(), oldPaths, "", "")
		if err != nil {
			t.Error(err)
		}
		if len(newPaths) != 1 {
			t.Errorf("expected 1 path, got %d", len(newPaths))
		}
		methods, ok := newPaths["/a"]
		if !ok {
			t.Error("missing path '/a'")
		}
		if len(methods) != 1 {
			t.Errorf("expected 1 method, got %d", len(methods))
		}
		if _, ok = methods["get"]; !ok {
			t.Error("missing method 'get'")
		}
		if len(allowedRefs) != 1 {
			t.Errorf("expected 1 reference, got %d", len(newPaths))
		}
		if _, ok := allowedRefs["A"]; !ok {
			t.Error("missing reference 'A'")
		}
	})
	t.Run("none", func(t *testing.T) {
		ladonClt.TokenPolicies = map[string][]string{}
		newPaths, allowedRefs, err := srv.getNewPathsByToken(context.Background(), oldPaths, "", "")
		if err != nil {
			t.Error(err)
		}
		if len(newPaths) != 0 {
			t.Errorf("expected 0 paths, got %d", len(newPaths))
		}
		if len(allowedRefs) != 0 {
			t.Errorf("expected 0 references, got %d", len(newPaths))
		}
	})
}

func TestService_transformDoc(t *testing.T) {
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

type ladonCltMock struct {
	RolePolicies  map[string]map[string]struct{}
	TokenPolicies map[string][]string
	Err           error
}

func (m *ladonCltMock) GetRoleAccessPolicy(_ context.Context, _, path, method string) (bool, error) {
	if m.Err != nil {
		return false, m.Err
	}
	methods, ok := m.RolePolicies[path]
	if !ok {
		return false, nil
	}
	_, ok = methods[method]
	return ok, nil
}

func (m *ladonCltMock) GetUserAccessPolicy(_ context.Context, _ string, _ map[string][]string) (map[string][]string, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.TokenPolicies, nil
}
