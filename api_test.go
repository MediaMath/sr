package sr

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAddError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errorObj := &Error{Code: 5555, Message: "Boom"}
		b, err := json.Marshal(&errorObj)
		if err != nil {
			t.Fatal(err)
		}

		http.Error(w, string(b), 500)
	}))
	defer ts.Close()

	host, err := NewHost(ts.URL, false)
	if err != nil {
		t.Fatal(err)
	}

	_, err = host.AddSchema("boo", &Schema{Schema: "boom"})
	if err == nil {
		t.Fatal("No error")
	}

}

func TestAdd(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, fmt.Sprintf("Wrong Method: %v", r.Method), 500)
		}

		if r.URL.Path != "/subjects/boo/versions" {
			http.Error(w, fmt.Sprintf("Wrong path: %v", r.URL.Path), 500)
		}

		id := &SchemaID{ID: 16}
		b, err := json.Marshal(&id)
		if err != nil {
			t.Fatal(err)
		}

		w.Write(b)
	}))
	defer ts.Close()

	host, err := NewHost(ts.URL, false)
	if err != nil {
		t.Fatal(err)
	}

	result, err := host.AddSchema("boo", &Schema{Schema: "boom"})
	if err != nil {
		t.Fatal(err)
	}

	if result.ID != 16 {
		t.Errorf("Wrong results: %v", result)
	}
}

func TestListSubjects(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, fmt.Sprintf("Wrong Method: %v", r.Method), 500)
		}

		if r.URL.Path != "/subjects" {
			http.Error(w, fmt.Sprintf("Wrong path: %v", r.URL.Path), 500)
		}

		subjects := []string{"boo", "goo"}
		b, err := json.Marshal(&subjects)
		if err != nil {
			t.Fatal(err)
		}

		w.Write(b)
	}))
	defer ts.Close()

	host, err := NewHost(ts.URL, false)
	if err != nil {
		t.Fatal(err)
	}

	result, err := host.ListSubjects()
	if err != nil {
		t.Fatal(err)
	}

	if result[0] != "boo" || result[1] != "goo" {
		t.Errorf("Wrong results: %v", result)
	}
}

func TestListVersions(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, fmt.Sprintf("Wrong Method: %v", r.Method), 500)
		}

		if r.URL.Path != "/subjects/goo/versions" {
			http.Error(w, fmt.Sprintf("Wrong path: %v", r.URL.Path), 500)
		}

		versions := []int{1, 4}
		b, err := json.Marshal(&versions)
		if err != nil {
			t.Fatal(err)
		}

		w.Write(b)
	}))
	defer ts.Close()

	host, err := NewHost(ts.URL, false)
	if err != nil {
		t.Fatal(err)
	}

	result, err := host.ListVersions("goo")
	if err != nil {
		t.Fatal(err)
	}

	if result[0] != 1 || result[1] != 4 {
		t.Errorf("Wrong results: %v", result)
	}

}

func TestGetVersion(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, fmt.Sprintf("Wrong Method: %v", r.Method), 500)
		}

		if r.URL.Path != "/subjects/goo/versions/8" {
			http.Error(w, fmt.Sprintf("Wrong path: %v", r.URL.Path), 500)
		}

		version := &Version{Version: 8, Name: "boom", Schema: "yeah"}
		b, err := json.Marshal(&version)
		if err != nil {
			t.Fatal(err)
		}

		w.Write(b)
	}))
	defer ts.Close()

	host, err := NewHost(ts.URL, false)
	if err != nil {
		t.Fatal(err)
	}

	result, err := host.GetVersion("goo", "8")
	if err != nil {
		t.Fatal(err)
	}

	if result.Version != 8 || result.Name != "boom" || result.Schema != "yeah" {
		t.Errorf("Wrong results: %v", result)
	}
}
