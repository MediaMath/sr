package sr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
)

//Host is the schema registry endpoint and is not concurrent safe
type Host struct {
	u       *url.URL
	verbose bool
}

//NewHost creates an api endpoint at the provided address
func NewHost(address string, verbose bool) (host *Host, err error) {
	var u *url.URL
	u, err = url.Parse(address)
	if err == nil {
		host = &Host{u, verbose}
	}

	return
}

//AddSchema adds a schema version to the provided subject
func (h *Host) AddSchema(subject string, schema *Schema) (id *SchemaID, err error) {
	id = &SchemaID{}
	err = h.post(path.Join("subjects", subject, "versions"), schema, id)
	return
}

//CheckSchema checks to see if a schema has already been registered for a subject
func (h *Host) CheckSchema(subject string, schema *Schema) (checked *CheckedSchema, err error) {
	checked = &CheckedSchema{}
	err = h.post(path.Join("subjects", subject), schema, checked)
	return
}

//CheckIsCompatible checks to see if a schema is compatible with a subject and version
func (h *Host) CheckIsCompatible(subject string, version string, schema *Schema) (is *IsCompatible, err error) {
	is = &IsCompatible{}
	err = h.post(path.Join("compatibility", "subjects", subject, "versions", version), schema, is)
	return
}

//GetSchema gets a schema by id
func (h *Host) GetSchema(id int) (schema *Schema, err error) {
	schema = &Schema{}
	err = h.get(path.Join("schemas", "ids", fmt.Sprintf("%v", id)), schema)
	return
}

//ListSubjects shows all the subjects that are registered
func (h *Host) ListSubjects() (result []string, err error) {
	result = []string{}
	err = h.get("subjects", &result)
	return
}

//ListVersions shows all the versions that are registered for a subject
func (h *Host) ListVersions(subject string) (result []int, err error) {
	result = []int{}
	err = h.get(path.Join("subjects", subject, "versions"), &result)
	return
}

//GetVersion shows a specific version version can be 'latest' or a positive integer
func (h *Host) GetVersion(subject string, version string) (result *Version, err error) {
	result = &Version{}
	err = h.get(path.Join("subjects", subject, "versions", version), result)
	return
}

const contentType = "application/vnd.schemaregistry.v1+json"

func (h *Host) parseResponse(resp *http.Response, err error, response interface{}) error {
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if h.verbose {
		log.Printf("Status: %v", resp.StatusCode)
		log.Printf("Header: %v", resp.Header)
		log.Printf("Body: %s", b)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		result := &Error{}
		err = json.Unmarshal(b, result)
		if err != nil {
			return err
		}

		return result
	}

	err = json.Unmarshal(b, response)
	return err
}

func (h *Host) post(path string, payload interface{}, response interface{}) error {
	h.u.Path = path

	p, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(h.u.String(), contentType, bytes.NewBuffer(p))
	return h.parseResponse(resp, err, response)
}

func (h *Host) get(path string, response interface{}) error {
	h.u.Path = path
	req, err := http.NewRequest("GET", h.u.String(), nil)
	if err != nil {
		return err
	}

	req.Header.Add("Accept", contentType)

	resp, err := http.DefaultClient.Do(req)
	return h.parseResponse(resp, err, response)
}
