package templateengine

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type StepOutput struct {
	ActionType string          `json:"action_type"`
	OrderIndex int             `json:"order_index"`
	Output     json.RawMessage `json:"output"`
}

// Matches {{...}} pattern
var templatePattern = regexp.MustCompile(`\{\{\s*(.+?)\s*\}\}`)

// Replaces any string values that contain {{ epxr }}
func Resolve(config map[string]any, payload []byte, steps []StepOutput) map[string]any {
	out := make(map[string]any, len(config))
	for k, v := range config {
		switch val := v.(type) {
		case string:
			out[k] = resolveString(val, payload, steps)
		default:
			out[k] = v
		}
	}
	return out
}

func resolveString(s string, payload []byte, steps []StepOutput) string {
	return templatePattern.ReplaceAllStringFunc(s, func(match string) string {
		inner := templatePattern.FindStringSubmatch(match)
		if len(inner) < 2 {
			return match
		}
		expr := strings.TrimSpace(inner[1])
		val, err := evaluate(expr, payload, steps)
		if err != nil {
			return match
		}
		return val
	})
}

func evaluate(expr string, payload []byte, steps []StepOutput) (string, error) {
	parts := strings.SplitN(expr, ".", 2)
	root := parts[0]

	switch {
	case root == "payload":
		if len(parts) == 1 {
			return string(payload), nil
		}
		return drillJSON(payload, parts[1])

	case root == "prev":
		if len(steps) == 0 {
			return "", fmt.Errorf("no previous step")
		}
		prev := steps[len(steps)-1]
		if len(parts) == 1 {
			return string(prev.Output), nil
		}
		rest := parts[1]
		rest = strings.TrimPrefix(rest, "output")
		rest = strings.TrimPrefix(rest, ".")
		if rest == "" {
			return string(prev.Output), nil
		}
		return drillJSON(prev.Output, rest)

	case strings.HasPrefix(root, "steps["):
		idx, err := parseStepIndex(root)
		if err != nil {
			return "", err
		}
		if idx < 0 || idx >= len(steps) {
			return "", fmt.Errorf("step index %d out of range (have %d steps)", idx, len(steps))
		}
		step := steps[idx]
		if len(parts) == 1 {
			return string(step.Output), nil
		}
		rest := parts[1]
		rest = strings.TrimPrefix(rest, "output")
		rest = strings.TrimPrefix(rest, ".")
		if rest == "" {
			return string(step.Output), nil
		}
		return drillJSON(step.Output, rest)

	default:
		return "", fmt.Errorf("unknown template root: %s", root)
	}
}

func parseStepIndex(s string) (int, error) {
	s = strings.TrimPrefix(s, "steps[")
	s = strings.TrimSuffix(s, "]")
	return strconv.Atoi(s)
}

func drillJSON(raw []byte, path string) (string, error) {
	if len(raw) == 0 {
		return "", fmt.Errorf("empty JSON")
	}
	var obj any
	if err := json.Unmarshal(raw, &obj); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}
	segments := strings.Split(path, ".")
	current := obj
	for _, seg := range segments {
		m, ok := current.(map[string]any)
		if !ok {
			return "", fmt.Errorf("cannot drill into non-object at %q", seg)
		}
		current, ok = m[seg]
		if !ok {
			return "", fmt.Errorf("key %q not found", seg)
		}
	}
	switch v := current.(type) {
	case string:
		return v, nil
	default:
		b, _ := json.Marshal(v)
		return string(b), nil
	}
}
