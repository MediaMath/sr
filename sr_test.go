package sr

//Copyright 2016 MediaMath <http://www.mediamath.com>.  All rights reserved.
//Use of this source code is governed by a BSD-style
//license that can be found in the LICENSE file.

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func tstClient() HTTPClient {
	return http.DefaultClient
}

func TestSchemaRegistryGetLatest(t *testing.T) {
	t.Skip("TODO: Race condition in this test")
	url := GetFunctionalTestURL(t)
	client := tstClient()

	toRegister := UniqueSchema()
	subject := UniqueSubject()

	id, schema, err := GetLatestSchema(client, url, subject)
	require.NotNil(t, err, fmt.Sprintf("Shouldn't be able to get a schema for an unregistered subject: %v %v %v", subject, id, schema))

	id1, err := Register(client, url, subject, toRegister)
	require.Nil(t, err, fmt.Sprintf("%v", err))
	assert.NotEqual(t, 0, id1)

	id2, schema, err := GetLatestSchema(client, url, subject)
	require.Nil(t, err, fmt.Sprintf("%v", err))
	assert.Equal(t, id1, id2)
	assert.Equal(t, toRegister, schema)
}

func TestSchemaRegistryRegisterCompatibleChange(t *testing.T) {
	url := GetFunctionalTestURL(t)
	client := tstClient()

	unique := time.Now().UnixNano()
	toRegister := TestSchema(unique)
	subject := UniqueSubject()

	id1, err := Register(client, url, subject, toRegister)
	require.Nil(t, err, fmt.Sprintf("%v", err))
	assert.NotEqual(t, uint32(0), id1)

	//This is a compatible change to the test Schema
	change := Schema(fmt.Sprintf(
		`{
	"namespace": "com.mediamath.sr",
	"type": "record",
	"name": "unit_test_functional_%v",
	"doc": "unit test schema with unique name",
	"fields": [
		{ "name": "foo", "type": "long", "doc": "a long for testing" },
		{ "name": "bar", "type": "string", "doc": "a string for testing"},
		{ "name": "bax", "type": ["null", "string"], "default": null, "doc": "a string for testing"}
	]
}
`, unique))

	id2, err := Register(client, url, subject, change)
	require.Nil(t, err, fmt.Sprintf("%v", err))
	assert.NotEqual(t, id1, id2)
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
			http.Error(w, err.Error(), 500)
		}

		w.Write(b)
	}))
	defer ts.Close()

	result, err := ListSubjects(tstClient(), ts.URL)
	require.Nil(t, err, fmt.Sprintf("%v", err))
	require.Equal(t, 2, len(result))
	assert.Equal(t, Subject("boo"), result[0])
	assert.Equal(t, Subject("goo"), result[1])
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
			http.Error(w, err.Error(), 500)
		}

		w.Write(b)
	}))
	defer ts.Close()

	result, err := ListVersions(tstClient(), ts.URL, Subject("goo"))
	require.Nil(t, err, fmt.Sprintf("%v", err))
	require.Equal(t, 2, len(result))
	assert.Equal(t, 1, result[0])
	assert.Equal(t, 4, result[1])
}

func TestGetVersion(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, fmt.Sprintf("Wrong Method: %v", r.Method), 500)
		}

		if r.URL.Path != "/subjects/goo/versions/8" {
			http.Error(w, fmt.Sprintf("Wrong path: %v", r.URL.Path), 500)
		}

		version := `{"version":8, "schema": "yeah", "subject":"goo", "id":19}`
		w.Write([]byte(version))
	}))
	defer ts.Close()

	id, schema, err := GetVersion(tstClient(), ts.URL, Subject("goo"), "8")
	require.Nil(t, err, fmt.Sprintf("%v", err))
	assert.Equal(t, uint32(19), id)
	assert.Equal(t, Schema("yeah"), schema)
}
