package templateengine

import (
	"encoding/json"
	"testing"
)

func TestResolve_PayloadSimple(t *testing.T) {
	cfg := map[string]any{
		"message": "Got: {{payload}}",
	}
	result := Resolve(cfg, []byte(`{"name":"amaan"}`), nil)
	want := `Got: {"name":"amaan"}`
	if result["message"] != want {
		t.Errorf("got %q, want %q", result["message"], want)
	}
}

func TestResolve_PayloadDrill(t *testing.T) {
	cfg := map[string]any{
		"greeting": "Hello {{payload.user.name}}!",
	}
	payload := []byte(`{"user":{"name":"Amaan"}}`)
	result := Resolve(cfg, payload, nil)
	if result["greeting"] != "Hello Amaan!" {
		t.Errorf("got %q", result["greeting"])
	}
}

func TestResolve_PrevOutput(t *testing.T) {
	steps := []StepOutput{
		{
			ActionType: "http_request",
			OrderIndex: 0,
			Output:     json.RawMessage(`{"status":"ok","count":42}`),
		},
	}
	cfg := map[string]any{
		"body": "Previous result: {{prev.output.status}}, count={{prev.output.count}}",
	}
	result := Resolve(cfg, nil, steps)
	if result["body"] != "Previous result: ok, count=42" {
		t.Errorf("got %q", result["body"])
	}
}

func TestResolve_PrevOutputFull(t *testing.T) {
	steps := []StepOutput{
		{
			ActionType: "http_request",
			OrderIndex: 0,
			Output:     json.RawMessage(`{"data":"hello"}`),
		},
	}
	cfg := map[string]any{
		"body": "{{prev.output}}",
	}
	result := Resolve(cfg, nil, steps)
	if result["body"] != `{"data":"hello"}` {
		t.Errorf("got %q", result["body"])
	}
}

func TestResolve_StepsIndex(t *testing.T) {
	steps := []StepOutput{
		{
			ActionType: "debug_log",
			OrderIndex: 0,
			Output:     json.RawMessage(`{"logged":true}`),
		},
		{
			ActionType: "http_request",
			OrderIndex: 1,
			Output:     json.RawMessage(`{"response":"pong"}`),
		},
	}
	cfg := map[string]any{
		"msg": "Step0={{steps[0].output.logged}}, Step1={{steps[1].output.response}}",
	}
	result := Resolve(cfg, nil, steps)
	if result["msg"] != "Step0=true, Step1=pong" {
		t.Errorf("got %q", result["msg"])
	}
}

func TestResolve_NoPrevSteps(t *testing.T) {
	cfg := map[string]any{
		"msg": "no prev: {{prev.output.x}}",
	}
	result := Resolve(cfg, nil, nil)
	if result["msg"] != "no prev: {{prev.output.x}}" {
		t.Errorf("got %q", result["msg"])
	}
}

func TestResolve_NonStringValuesUntouched(t *testing.T) {
	cfg := map[string]any{
		"count":   42,
		"enabled": true,
		"message": "hello {{payload.name}}",
	}
	payload := []byte(`{"name":"world"}`)
	result := Resolve(cfg, payload, nil)
	if result["count"] != 42 {
		t.Errorf("count should be untouched, got %v", result["count"])
	}
	if result["enabled"] != true {
		t.Errorf("enabled should be untouched, got %v", result["enabled"])
	}
	if result["message"] != "hello world" {
		t.Errorf("message got %q", result["message"])
	}
}

func TestResolve_OutOfBoundsStepIndex(t *testing.T) {
	cfg := map[string]any{
		"msg": "{{steps[5].output.x}}",
	}
	result := Resolve(cfg, nil, nil)
	if result["msg"] != "{{steps[5].output.x}}" {
		t.Errorf("got %q", result["msg"])
	}
}

func TestResolve_MultipleTemplatesInOneString(t *testing.T) {
	steps := []StepOutput{
		{Output: json.RawMessage(`{"a":"alpha"}`)},
	}
	cfg := map[string]any{
		"msg": "payload={{payload.x}} prev={{prev.output.a}}",
	}
	result := Resolve(cfg, []byte(`{"x":"hello"}`), steps)
	if result["msg"] != "payload=hello prev=alpha" {
		t.Errorf("got %q", result["msg"])
	}
}
