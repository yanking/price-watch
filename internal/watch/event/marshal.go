package event

import (
	"encoding/json"
	"fmt"
)

func Marshal[T any](e Event[T]) ([]byte, error) {
	return json.Marshal(e)
}

func Unmarshal(data []byte) (eventType string, raw Event[json.RawMessage], err error) {
	if err = json.Unmarshal(data, &raw); err != nil {
		return "", raw, fmt.Errorf("unmarshal event envelope: %w", err)
	}
	return raw.Type, raw, nil
}

func UnmarshalData[T any](raw json.RawMessage) (T, error) {
	var v T
	if err := json.Unmarshal(raw, &v); err != nil {
		return v, fmt.Errorf("unmarshal event data: %w", err)
	}
	return v, nil
}
