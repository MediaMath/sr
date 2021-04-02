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
	"strings"
)

//Schema is a string that represents a avro schema
type Schema string

//EmptySchema is the 'zero' value for a Schema
var EmptySchema = Schema("")

//SchemaJSON is what the schema registry expects when sending it a schema
type SchemaJSON struct {
	Schema Schema `json:"schema"`
}

//ConfigGetJSON is what the schema registry returns on config get endpoints
type ConfigGetJSON struct {
	Compatibility string `json:"compatibilityLevel"`
}

//ConfigPutJSON is what the schema registry expects on config put endpoints
type ConfigPutJSON struct {
	Compatibility string `json:"compatibility"`
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
	if err == nil {
		isCompatible := struct {
			IsCompatible bool `json:"is_compatible"`
		}{}

		var status int
		var body []byte
		status, body, err = doJSON(client, req, &isCompatible)
		if status != 200 {
			err = fmt.Errorf("Unexpected return code: %v:%s", status, body)
		}

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

//GetSubjectDerivedCompatibility returns the compatibility level for a subject or the default if a subject specific doesnt exist
func GetSubjectDerivedCompatibility(client HTTPClient, url string, subject Subject) (compatibility Compatibility, err error) {
	compatibility = Zero

	var status int
	var req *http.Request
	req, err = GetSubjectConfigRequest(url, subject)
	if err == nil {
		status, compatibility, err = compatibilityJSON(client, req)
	}

	if err == nil && status == http.StatusNotFound {
		compatibility, err = GetDefaultCompatibility(client, url)
	}

	return
}

//SetSubjectCompatibility sets the compatibility level for a subject
func SetSubjectCompatibility(client HTTPClient, url string, subject Subject, compatibility Compatibility) (result Compatibility, err error) {
	result = Zero

	var (
		req          *http.Request
		responseBody []byte
		status       int

		body     = &ConfigPutJSON{Compatibility: string(compatibility)}
		response = &ConfigPutJSON{}
	)

	req, err = PutSubjectConfigRequest(url, subject, body)
	if err == nil {
		status, responseBody, err = doJSON(client, req, response)
	}

	if err == nil && status != http.StatusOK {
		err = fmt.Errorf("Unknown response (%v) (%s)", status, responseBody)
	}

	if err == nil {
		result = Compatibility(response.Compatibility)
	}

	return
}

//GetSubjectCompatibility returns the compatibility level for a subject
func GetSubjectCompatibility(client HTTPClient, url string, subject Subject) (compatibility Compatibility, err error) {
	compatibility = Zero

	var req *http.Request
	req, err = GetSubjectConfigRequest(url, subject)
	if err == nil {
		_, compatibility, err = compatibilityJSON(client, req)
	}

	return
}

//GetDefaultCompatibility returns the compatibility level set at the server level
func GetDefaultCompatibility(client HTTPClient, url string) (compatibility Compatibility, err error) {
	compatibility = Zero

	var req *http.Request
	req, err = GetConfigRequest(url)
	if err == nil {
		_, compatibility, err = compatibilityJSON(client, req)
	}

	return
}

func compatibilityJSON(client HTTPClient, req *http.Request) (status int, compatibility Compatibility, err error) {
	compatibility = Zero

	configResponse := &ConfigGetJSON{}
	status, _, err = doJSON(client, req, &configResponse)

	if err == nil {
		compatibility = Compatibility(configResponse.Compatibility)
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

//GetConfigRequest returns the http.Request for the GET /config route
func GetConfigRequest(baseURL string) (*http.Request, error) {
	return get(baseURL, "config")
}

//GetSubjectConfigRequest returns the http.Request for the GET /config route
func GetSubjectConfigRequest(baseURL string, subject Subject) (*http.Request, error) {
	return get(baseURL, path.Join("config", string(subject)))
}

//PutSubjectConfigRequest returns the http.Request for the Put /config/<subject> route
func PutSubjectConfigRequest(baseURL string, subject Subject, body *ConfigPutJSON) (*http.Request, error) {
	return put(baseURL, path.Join("config", string(subject)), body)
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

func put(baseURL, query string, body interface{}) (request *http.Request, err error) {
	return putOrPost(baseURL, "PUT", query, body)
}

func post(baseURL, query string, body interface{}) (request *http.Request, err error) {
	return putOrPost(baseURL, "POST", query, body)
}

func putOrPost(baseURL, method string, query string, body interface{}) (request *http.Request, err error) {
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

	request, err = http.NewRequest(method, u, reader)
	if request != nil {
		request.Header.Add("Accept", schemaRegistryAccepts)
		request.Header.Add("Content-Type", "application/vnd.schemaregistry.v1+json")
	}

	return
}

func buildURL(baseURL, endpoint string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	u.Path = path.Join(u.Path, endpoint)
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

		if err != nil {
			err = fmt.Errorf("Unexpected response (%v) from %v.\n%s", status, request.URL, body)
		}
	}

	return
}

func Copy(client HTTPClient, fromURL, toURL, fromPrefix, toPrefix string) (int, error) {
	var total int
	var subjects, err = ListSubjects(client, fromURL)
	if err != nil {
		return total, err
	}

	for _, subject := range subjects {
		var _, schema, err = GetLatestSchema(client, fromURL, subject)
		if err != nil {
			return total, err
		}

		var toSubject = strings.Replace(string(subject), fromPrefix, toPrefix, 1)
		_, err = Register(client, toURL, Subject(toSubject), schema)
		if err != nil {
			return total, err
		}

		total++
	}

	return total, nil
}
