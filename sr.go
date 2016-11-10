package sr

//Copyright 2016 MediaMath <http://www.mediamath.com>.  All rights reserved.
//Use of this source code is governed by a BSD-style
//license that can be found in the LICENSE file.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
)

//Schema is a string that represents a avro schema
type Schema string

//EmptySchema is the 'zero' value for a Schema
var EmptySchema = Schema("")

//SchemaJSON is what the schema registry expects when sending it a schema
type SchemaJSON struct {
	Schema Schema `json:"schema"`
}

//GetLatestSchema returns the latest schema and id for a subject
func GetLatestSchema(client HTTPClient, url string, subject Subject) (id uint32, schema Schema, err error) {
	return GetVersion(client, url, subject, "latest")
}

//GetVersion returns a schema and id for a subject and version
func GetVersion(client HTTPClient, url string, subject Subject, version string) (id uint32, schema Schema, err error) {
	schema = EmptySchema

	var req *http.Request
	var status int
	var body []byte

	req, err = GetVersionRequest(url, subject, version)
	if err == nil {

		schemaResponse := struct {
			ID     uint32 `json:"id"`
			Schema Schema `json:"schema"`
		}{}
		status, body, err = doJSON(client, req, &schemaResponse)

		if err == nil {
			id = schemaResponse.ID
			schema = schemaResponse.Schema
		}
	}

	if err == nil && schema == EmptySchema {
		err = fmt.Errorf("%v:%s", status, body)
	}

	return
}

//GetSchema returns a schema for an id
func GetSchema(client HTTPClient, url string, id uint32) (schema Schema, err error) {
	schema = EmptySchema

	var req *http.Request
	req, err = GetSchemaRequest(url, id)
	if err == nil {

		schemaResponse := &SchemaJSON{}
		_, _, err = doJSON(client, req, &schemaResponse)

		if err == nil {
			schema = schemaResponse.Schema
		}
	}

	return
}

//Register adds a schema to a subject and returns the new id
func Register(client HTTPClient, url string, subject Subject, schema Schema) (id uint32, err error) {

	var req *http.Request
	var status int
	var result []byte

	body := SchemaJSON{schema}
	req, err = RegisterRequest(url, subject, &body)
	if err == nil {

		idResponse := struct {
			ID uint32 `json:"id"`
		}{}

		status, result, err = doJSON(client, req, &idResponse)

		if err == nil {
			id = idResponse.ID
		}
	}

	if err == nil && id == uint32(0) {
		err = fmt.Errorf("%v:%v:%s", status, req, result)
	}

	return
}

//HasSchema returns the version and id for a schema on a subject
func HasSchema(client HTTPClient, url string, subject Subject, schema Schema) (version int, id int, err error) {
	var req *http.Request
	body := &SchemaJSON{schema}
	req, err = HasSchemaRequest(url, subject, body)
	if err == nil {

		checkedSchema := struct {
			Schema  Schema  `json:"schema"`
			Version int     `json:"version"`
			Subject Subject `json:"subject"`
			ID      int     `json:"id"`
		}{}

		_, _, err = doJSON(client, req, &checkedSchema)
		if err == nil {
			version = checkedSchema.Version
			id = checkedSchema.ID
		}
	}

	return
}

//IsCompatible will return if the provided schema is compatible with the subject and version provided. Version can either be a numeric version or 'latest'
func IsCompatible(client HTTPClient, url string, subject Subject, version string, schema Schema) (is bool, err error) {
	var req *http.Request
	body := &SchemaJSON{schema}
	req, err = CheckIsCompatibleRequest(url, subject, version, body)
	if err != nil {

		isCompatible := struct {
			IsCompatible bool `json:"is_compatible"`
		}{}

		_, _, err = doJSON(client, req, &isCompatible)
		if err == nil {
			is = isCompatible.IsCompatible
		}
	}

	return
}

//ListSubjects returns the list of subjects
func ListSubjects(client HTTPClient, url string) (subjects []Subject, err error) {
	var req *http.Request
	req, err = ListSubjectsRequest(url)
	if err == nil {
		_, _, err = doJSON(client, req, &subjects)
	}

	return
}

//ListVersions returns the list of versions for a subject
func ListVersions(client HTTPClient, url string, subject Subject) (versions []int, err error) {
	var req *http.Request
	req, err = ListVersionsRequest(url, subject)
	if err == nil {
		_, _, err = doJSON(client, req, &versions)
	}

	return
}

//GetSchemaRequest returns the http.Request for GET /schemas/ids/<id> route
func GetSchemaRequest(baseURL string, id uint32) (*http.Request, error) {
	return get(baseURL, path.Join("schemas", "ids", fmt.Sprintf("%v", id)))
}

//RegisterRequest returns the http.Request for the POST  /subjects/<subject>/versions
func RegisterRequest(baseURL string, subject Subject, body *SchemaJSON) (*http.Request, error) {
	return post(baseURL, path.Join("subjects", string(subject), "versions"), body)
}

//GetVersionRequest returns the http.Request for the GET /subjects/<subject>/versions/<version> version can either be a number or 'latest'
func GetVersionRequest(baseURL string, subject Subject, version string) (*http.Request, error) {
	return get(baseURL, path.Join("subjects", string(subject), "versions", version))
}

//HasSchemaRequest returns the http.Request for the POST /subjects/<subject>
func HasSchemaRequest(baseURL string, subject Subject, body *SchemaJSON) (*http.Request, error) {
	return post(baseURL, path.Join("subjects", string(subject)), body)
}

//CheckIsCompatibleRequest returns the http.Request for the POST /compatibility/subjects/<subject>/versions/<version> route
func CheckIsCompatibleRequest(baseURL string, subject Subject, version string, body *SchemaJSON) (*http.Request, error) {
	return post(baseURL, path.Join("compatibility", "subjects", string(subject), "versions", version), body)
}

//ListSubjectsRequest returns the GET /subjects
func ListSubjectsRequest(baseURL string) (*http.Request, error) {
	return get(baseURL, "subjects")
}

//ListVersionsRequest returns GET /subjects/<subject>/versions
func ListVersionsRequest(baseURL string, subject Subject) (*http.Request, error) {
	return get(baseURL, path.Join("subjects", string(subject), "versions"))
}

const schemaRegistryAccepts = "application/vnd.schemaregistry.v1+json,application/vnd.schemaregistry+json, application/json"

func get(baseURL, query string) (request *http.Request, err error) {
	var u string
	u, err = buildURL(baseURL, query)
	if err != nil {
		return
	}

	request, err = http.NewRequest("GET", u, nil)
	if request != nil {
		request.Header.Add("Accept", schemaRegistryAccepts)
	}

	return
}

func post(baseURL, query string, body interface{}) (request *http.Request, err error) {
	var reader io.Reader
	if body != nil {
		var data []byte
		data, err = json.Marshal(body)
		if err != nil {
			return
		}
		reader = bytes.NewBuffer(data)
	}

	var u string
	u, err = buildURL(baseURL, query)
	if err != nil {
		return
	}

	request, err = http.NewRequest("POST", u, reader)
	if request != nil {
		request.Header.Add("Accept", schemaRegistryAccepts)
		request.Header.Add("Content-Type", "application/vnd.schemaregistry.v1+json")
	}

	return
}

func buildURL(baseURL, path string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	u.Path += path
	return u.String(), nil
}

//HTTPClient is any client that can do a http request
type HTTPClient interface {
	Do(request *http.Request) (*http.Response, error)
}

func doJSON(restful HTTPClient, request *http.Request, response interface{}) (status int, body []byte, err error) {
	res, err := restful.Do(request)
	if err == nil {
		body, err = ioutil.ReadAll(res.Body)
		res.Body.Close()
		status = res.StatusCode
	}

	if err == nil && response != nil {
		err = json.Unmarshal(body, response)
	}

	return
}
