package sr

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"path"
)

//Host is the schema registry endpoint and is not concurrent safe
type Host struct {
	u *url.URL
}

//NewHost creates an api endpoint at the provided address
func NewHost(address string) (host *Host, err error) {
	var u *url.URL
	u, err = url.Parse(address)
	if err == nil {
		host = &Host{u}
	}

	return
}

//AddSchema adds a schema version to the provided subject
func (h *Host) AddSchema(subject string, schema *Schema) (*http.Response, error) {
	return h.post(path.Join("subjects", subject, "versions"), schema)
}

//ListSubjects shows all the subjects that are registered
func (h *Host) ListSubjects() (*http.Response, error) {
	return h.get("subjects")
}

//ListVersions shows all the versions that are registered for a subject
func (h *Host) ListVersions(subject string) (*http.Response, error) {
	return h.get(path.Join("subjects", subject, "versions"))
}

//GetVersion shows a specific version version can be 'latest' or a positive integer
func (h *Host) GetVersion(subject string, version string) (*http.Response, error) {
	return h.get(path.Join("subjects", subject, "versions", version))
}

const contentType = "application/vnd.schemaregistry.v1+json"

func (h *Host) post(path string, payload interface{}) (*http.Response, error) {
	h.u.Path = path

	p, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return http.Post(h.u.String(), contentType, bytes.NewBuffer(p))
}

func (h *Host) get(path string) (*http.Response, error) {
	h.u.Path = path
	req, err := http.NewRequest("GET", h.u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", contentType)

	return http.DefaultClient.Do(req)
}
