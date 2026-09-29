package request

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

const SchemaVersion = "0.1.0"

func BuildPrompt(prompt string, timeoutSeconds int) ([]byte, error) {
	if strings.TrimSpace(prompt) == "" {
		return nil, errors.New("prompt must not be empty")
	}
	executionID, err := identifier("exec")
	if err != nil {
		return nil, err
	}
	correlationID, err := identifier("corr")
	if err != nil {
		return nil, err
	}
	body := map[string]interface{}{
		"schemaVersion":     SchemaVersion,
		"executionId":       executionID,
		"correlationId":     correlationID,
		"harness":           "pi.dev",
		"task":              prompt,
		"timeoutSeconds":    timeoutSeconds,
		"contextPacketRefs": []string{},
		"toolGrants":        []string{},
	}
	return marshal(body)
}

func LoadFile(path string) ([]byte, error) {
	if path == "" {
		return nil, errors.New("request path must not be empty")
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read request %s: %w", path, err)
	}
	if err := Validate(body); err != nil {
		return nil, fmt.Errorf("invalid request %s: %w", path, err)
	}
	return body, nil
}

func Validate(body []byte) error {
	var object map[string]interface{}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err := decoder.Decode(&object); err != nil {
		return fmt.Errorf("request must be valid JSON: %w", err)
	}
	if object == nil {
		return errors.New("request must be a JSON object")
	}
	var extra interface{}
	if err := decoder.Decode(&extra); err != io.EOF {
		return errors.New("request must contain one JSON value")
	}
	stringField := func(name string) error {
		value, ok := object[name].(string)
		if !ok || strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
		return nil
	}
	if err := stringField("schemaVersion"); err != nil {
		return err
	}
	if object["schemaVersion"] != SchemaVersion {
		return fmt.Errorf("schemaVersion must be %s", SchemaVersion)
	}
	for _, field := range []string{"executionId", "correlationId", "harness", "task"} {
		if err := stringField(field); err != nil {
			return err
		}
	}
	if object["harness"] != "pi.dev" {
		return errors.New("harness must be pi.dev")
	}
	timeout, ok := object["timeoutSeconds"].(json.Number)
	if !ok {
		return errors.New("timeoutSeconds must be a positive integer")
	}
	parsed, err := timeout.Int64()
	if err != nil || parsed < 1 {
		return errors.New("timeoutSeconds must be a positive integer")
	}
	return nil
}

func Pretty(body []byte) ([]byte, error) {
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, body, "", "  "); err != nil {
		return nil, err
	}
	pretty.WriteByte('\n')
	return pretty.Bytes(), nil
}

func identifier(prefix string) (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate execution identifier: %w", err)
	}
	return prefix + "-" + hex.EncodeToString(raw[:]), nil
}

func marshal(value interface{}) ([]byte, error) {
	body, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("encode request: %w", err)
	}
	body = append(body, '\n')
	return body, nil
}
