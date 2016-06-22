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

func (h *Host) post(path string, payload interface{}) (*http.Response, error) {
	h.u.Path = path

	p, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return http.Post(h.u.String(), "application/vnd.schemaregistry.v1+json", bytes.NewBuffer(p))
}
