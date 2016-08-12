package sr

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	//TestURLEnvVar is the url to run functional tests against
	TestURLEnvVar = "SR_TEST_SCHEMA_REGISTRY"
	//TestRequiredEnvVar if set to true will make tests fail
	TestRequiredEnvVar = "SR_TEST_REQUIRED"
)

//IsFunctionalTestRequired returns whether SR_TEST_REQUIRED is set
func IsFunctionalTestRequired() bool {
	return strings.TrimSpace(os.Getenv(TestRequiredEnvVar)) == "true"
}

//HandleFunctionalTestError will skip or fail based on whether SR_TEST_REQUIRED is set
func HandleFunctionalTestError(t testing.TB, err error) {
	if err != nil && IsFunctionalTestRequired() {
		require.FailNow(t, err.Error())
	} else if err != nil {
		t.Skip(err)
	}
}

//UniqueSchema returns a schema with a unique name
func UniqueSchema() Schema {
	return TestSchema(time.Now().UnixNano())

}

//UniqueSubject returns a subject with a unique name
func UniqueSubject() Subject {
	unique := time.Now().Unix()
	return Subject(fmt.Sprintf("ut-%v", unique))
}

//TestSchema returns a schema with the unique part added to the name
func TestSchema(unique int64) Schema {
	return Schema(fmt.Sprintf(
		`{
	"namespace": "com.mediamath.sr",
	"type": "record",
	"name": "unit_test_functional_%v",
	"doc": "unit test schema with unique name",
	"fields": [
		{ "name": "foo", "type": "long", "doc": "a long for testing" },
		{ "name": "bar", "type": "string", "doc": "a string for testing"}
	]
}
`, unique))
}

//GetFunctionalTestURL skips, fails, or returns the config variable passed in
func GetFunctionalTestURL(t *testing.T) string {
	if testing.Short() {
		t.Skipf("Skipping %v tests in short mode", TestURLEnvVar)
	}

	value := strings.TrimSpace(os.Getenv(TestURLEnvVar))

	if value == "" {
		HandleFunctionalTestError(t, fmt.Errorf("%v is undefined", TestURLEnvVar))
	}

	return value
}
