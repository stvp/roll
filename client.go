package roll

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash/adler32"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strings"
	"time"
)

const (
	// By default, all Rollbar API requests are sent to this endpoint.
	endpoint = "https://api.rollbar.com/api/1/item/"

	// Identify this Rollbar client library to the Rollbar API.
	clientName     = "go-roll"
	clientVersion  = "0.0.1"
	clientLanguage = "go"
)

var (
	// Rollbar access token for the global client. If this is blank, no items
	// will be sent to Rollbar.
	Token = ""

	// Environment for all items reported with the global client.
	Environment = "development"
)

type rollbarSuccess struct {
	Result map[string]string `json:"result"`
}

// Client reports items to a single Rollbar project.
type Client interface {
	Critical(err error, custom map[string]string) (uuid string, e error)
	Error(err error, custom map[string]string) (uuid string, e error)
	Warning(err error, custom map[string]string) (uuid string, e error)
	Info(msg string, custom map[string]string) (uuid string, e error)
	Debug(msg string, custom map[string]string) (uuid string, e error)
}

type rollbarClient struct {
	token string
	env   string
}

// New creates a new Rollbar client that reports items to the given project
// token and with the given environment (eg. "production", "development", etc).
func New(token, env string) Client {
	return &rollbarClient{token, env}
}

func Critical(err error, custom map[string]string) (uuid string, e error) {
	client := rollbarClient{Token, Environment}
	return client.skipStack("critical", err, 3, custom)
}

func Error(err error, custom map[string]string) (uuid string, e error) {
	client := rollbarClient{Token, Environment}
	return client.skipStack("error", err, 3, custom)
}

func Warning(err error, custom map[string]string) (uuid string, e error) {
	client := rollbarClient{Token, Environment}
	return client.skipStack("warning", err, 3, custom)
}

func Info(msg string, custom map[string]string) (uuid string, e error) {
	return New(Token, Environment).Info(msg, custom)
}

func Debug(msg string, custom map[string]string) (uuid string, e error) {
	return New(Token, Environment).Debug(msg, custom)
}

func (c *rollbarClient) Critical(err error, custom map[string]string) (uuid string, e error) {
	return c.skipStack("critical", err, 3, custom)
}

func (c *rollbarClient) Error(err error, custom map[string]string) (uuid string, e error) {
	return c.skipStack("error", err, 3, custom)
}

func (c *rollbarClient) Warning(err error, custom map[string]string) (uuid string, e error) {
	return c.skipStack("warning", err, 3, custom)
}

func (c *rollbarClient) Info(msg string, custom map[string]string) (uuid string, e error) {
	item := c.buildMessageItem("info", msg, custom)
	return c.send(item)
}

func (c *rollbarClient) Debug(msg string, custom map[string]string) (uuid string, e error) {
	item := c.buildMessageItem("debug", msg, custom)
	return c.send(item)
}

func (c *rollbarClient) skipStack(level string, err error, skip int, custom map[string]string) (uuid string, e error) {
	item := c.buildTraceItem(level, err, buildStack(skip), custom)
	return c.send(item)
}

func (c *rollbarClient) buildTraceItem(level string, err error, s stack, custom map[string]string) (item map[string]interface{}) {
	item = c.buildItem(level, err.Error(), custom)
	itemData := item["data"].(map[string]interface{})
	itemData["fingerprint"] = stackFingerprint(err.Error(), s)
	itemData["body"] = map[string]interface{}{
		"trace": map[string]interface{}{
			"frames": s,
			"exception": map[string]interface{}{
				"class":   errorClass(err),
				"message": err.Error(),
			},
		},
	}

	return item
}

func (c *rollbarClient) buildMessageItem(level string, msg string, custom map[string]string) (item map[string]interface{}) {
	item = c.buildItem(level, msg, custom)
	itemData := item["data"].(map[string]interface{})
	itemData["body"] = map[string]interface{}{
		"message": map[string]interface{}{
			"body": msg,
		},
	}

	return item
}

func (c *rollbarClient) buildItem(level, title string, custom map[string]string) map[string]interface{} {
	hostname, _ := os.Hostname()

	return map[string]interface{}{
		"access_token": c.token,
		"data": map[string]interface{}{
			"environment": c.env,
			"title":       title,
			"level":       level,
			"timestamp":   time.Now().Unix(),
			"platform":    runtime.GOOS,
			"language":    clientLanguage,
			"server": map[string]interface{}{
				"host": hostname,
			},
			"notifier": map[string]interface{}{
				"name":    clientName,
				"version": clientVersion,
			},
			"custom": custom,
		},
	}
}

// send reports the given item to Rollbar and returns either a UUID for the
// reported item or an error.
func (c *rollbarClient) send(item map[string]interface{}) (uuid string, err error) {
	if len(c.token) == 0 {
		return "", nil
	}

	jsonBody, err := json.Marshal(item)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(endpoint, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return "", err
	}
	defer func() { resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println(string(body))
		return "", fmt.Errorf("Rollbar returned %s", resp.Status)
	}

	// Extract UUID from JSON response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", nil
	}
	success := rollbarSuccess{}
	json.Unmarshal(body, &success)

	return success.Result["uuid"], nil
}

// errorClass returns a class name for an error (eg.  "ErrUnexpectedEOF").  For
// string errors, it returns a checksum of the error string.
func errorClass(err error) string {
	class := reflect.TypeOf(err).String()
	if class == "" {
		return "panic"
	} else if class == "*errors.errorString" {
		checksum := adler32.Checksum([]byte(err.Error()))
		return fmt.Sprintf("{%x}", checksum)
	} else {
		return strings.TrimPrefix(class, "*")
	}
}
