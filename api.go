package sr

import (
	"bytes"
	"encoding/json"
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

func (h *Host) post(path string, payload interface{}, response interface{}) error {
	h.u.Path = path

	p, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(h.u.String(), contentType, bytes.NewBuffer(p))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if h.verbose {
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

func (h *Host) get(path string) (*http.Response, error) {
	h.u.Path = path
	req, err := http.NewRequest("GET", h.u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", contentType)

	return http.DefaultClient.Do(req)
}
