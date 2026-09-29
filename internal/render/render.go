package render

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

func JSON(w io.Writer, body []byte) error {
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, body, "", "  "); err != nil {
		return fmt.Errorf("response was not valid JSON: %w", err)
	}
	pretty.WriteByte('\n')
	_, err := w.Write(pretty.Bytes())
	return err
}

func HumanResult(w io.Writer, result map[string]interface{}) error {
	status, _ := result["status"].(string)
	if status == "" {
		status = "unknown"
	}
	if _, err := fmt.Fprintf(w, "Status: %s\n", status); err != nil {
		return err
	}
	for _, field := range []string{"executionId", "jobId", "runId", "assignmentId"} {
		if value, ok := result[field].(string); ok && value != "" {
			if _, err := fmt.Fprintf(w, "%s: %s\n", field, value); err != nil {
				return err
			}
		}
	}
	if output, ok := result["output"].(map[string]interface{}); ok {
		if text, ok := output["text"].(string); ok && text != "" {
			if _, err := fmt.Fprintf(w, "\n%s\n", text); err != nil {
				return err
			}
		} else if len(output) > 0 {
			encoded, _ := json.MarshalIndent(output, "", "  ")
			if _, err := fmt.Fprintf(w, "\nOutput:\n%s\n", encoded); err != nil {
				return err
			}
		}
	}
	if failure, ok := result["failure"].(map[string]interface{}); ok {
		category, _ := failure["category"].(string)
		message, _ := failure["message"].(string)
		if category != "" || message != "" {
			_, err := fmt.Fprintf(w, "\nFailure: %s: %s\n", category, message)
			return err
		}
	}
	return nil
}
