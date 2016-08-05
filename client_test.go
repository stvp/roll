package roll

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/pkg/errors"
)

// -- Test helpers

type CustomError struct {
	s string
}

func (e *CustomError) Error() string {
	return e.s
}

func setup() func() {
	Token = os.Getenv("TOKEN")
	Environment = "test"

	if Token == "" {
		Token = "test"
		originalEndpoint := DefaultEndpoint
		server := httptest.NewServer(http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{"result": {"uuid": "01234567890123456789012345678901"}}`))
			},
		))

		DefaultEndpoint = server.URL

		return func() {
			DefaultEndpoint = originalEndpoint
			server.Close()
		}
	}

	// Assume Token was provided and we want integration tests.
	return func() {}
}

// -- Tests

func TestErrorClass(t *testing.T) {
	errors := map[string]error{
		"{508e076d}":       fmt.Errorf("Something is broken!"),
		"roll.CustomError": &CustomError{"Terrible mistakes were made."},
	}

	for expected, err := range errors {
		if errorClass(err) != expected {
			t.Error("Got:", errorClass(err), "Expected:", expected)
		}
	}
}

func TestCritical(t *testing.T) {
	teardown := setup()
	defer teardown()

	uuid, err := Critical(errors.New("global critical"), map[string]string{"extras": "true"})
	if err != nil {
		t.Error(err)
	}
	if len(uuid) != 32 {
		t.Errorf("expected UUID, got: %#v", uuid)
	}
}

func TestError(t *testing.T) {
	teardown := setup()
	defer teardown()

	uuid, err := Error(errors.New("global error"), map[string]string{"extras": "true"})
	if err != nil {
		t.Error(err)
	}
	if len(uuid) != 32 {
		t.Errorf("expected UUID, got: %#v", uuid)
	}
}

func TestWarning(t *testing.T) {
	teardown := setup()
	defer teardown()

	uuid, err := Warning(errors.New("global warning"), map[string]string{"extras": "true"})
	if err != nil {
		t.Error(err)
	}
	if len(uuid) != 32 {
		t.Errorf("expected UUID, got: %#v", uuid)
	}
}

func TestInfo(t *testing.T) {
	teardown := setup()
	defer teardown()

	uuid, err := Info("global info", map[string]string{"extras": "true"})
	if err != nil {
		t.Error(err)
	}
	if len(uuid) != 32 {
		t.Errorf("expected UUID, got: %#v", uuid)
	}
}

func TestDebug(t *testing.T) {
	teardown := setup()
	defer teardown()

	uuid, err := Debug("global debug", map[string]string{"extras": "true"})
	if err != nil {
		t.Error(err)
	}
	if len(uuid) != 32 {
		t.Errorf("expected UUID, got: %#v", uuid)
	}
}

func TestRollbarClientCritical(t *testing.T) {
	teardown := setup()
	defer teardown()

	client := New(Token, Environment)

	uuid, err := client.Critical(errors.New("new client critical"), map[string]string{"extras": "true"})
	if err != nil {
		t.Error(err)
	}
	if len(uuid) != 32 {
		t.Errorf("expected UUID, got: %#v", uuid)
	}
}

func TestRollbarClientError(t *testing.T) {
	teardown := setup()
	defer teardown()

	client := New(Token, Environment)

	uuid, err := client.Error(errors.New("new client error"), map[string]string{"extras": "true"})
	if err != nil {
		t.Error(err)
	}
	if len(uuid) != 32 {
		t.Errorf("expected UUID, got: %#v", uuid)
	}
}

func TestRollbarClientWarning(t *testing.T) {
	teardown := setup()
	defer teardown()

	client := New(Token, Environment)

	uuid, err := client.Warning(errors.New("new client warning"), map[string]string{"extras": "true"})
	if err != nil {
		t.Error(err)
	}
	if len(uuid) != 32 {
		t.Errorf("expected UUID, got: %#v", uuid)
	}
}

func TestRollbarClientInfo(t *testing.T) {
	teardown := setup()
	defer teardown()

	client := New(Token, Environment)

	uuid, err := client.Info("new client info", map[string]string{"extras": "true"})
	if err != nil {
		t.Error(err)
	}
	if len(uuid) != 32 {
		t.Errorf("expected UUID, got: %#v", uuid)
	}
}

func TestRollbarClientDebug(t *testing.T) {
	teardown := setup()
	defer teardown()

	client := New(Token, Environment)

	uuid, err := client.Debug("new client debug", map[string]string{"extras": "true"})
	if err != nil {
		t.Error(err)
	}
	if len(uuid) != 32 {
		t.Errorf("expected UUID, got: %#v", uuid)
	}
}

func TestAssembleStackWrapped(t *testing.T) {
	err := errors.Wrap(errors.New("foo bar"), "fooing bar")
	client := &rollbarClient{}
	item := client.assembleStack(ERR, err, 3, nil)
	body, jerr := json.Marshal(item)
	if jerr != nil {
		t.Fatalf("error while marshaling to json: %q", jerr)
	}

	if !strings.Contains(string(body), "TestAssembleStackWrapped") {
		t.Errorf("expected TestAssembleStackWrapped in frames, but was not found.")
	}
}

func TestAssembleStackNormal(t *testing.T) {
	err := fmt.Errorf("foo bar")
	client := &rollbarClient{}
	item := client.assembleStack(ERR, err, 3, nil)
	body, jerr := json.Marshal(item)
	if jerr != nil {
		t.Fatalf("error while marshaling to json: %q", jerr)
	}

	if strings.Contains(string(body), "TestAssembleStackNormal") {
		t.Errorf("expected TestAssembleStackNormal to not be found in frames, but was.")
	}
}
