package lsif

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

func UnmarshalHoverData(element Element) (string, error) {
	type HoverResult struct {
		Contents json.RawMessage `json:"contents"`
	}
	type HoverVertex struct {
		Result HoverResult `json:"result"`
	}

	var payload HoverVertex
	if err := json.Unmarshal(element.Raw, &payload); err != nil {
		return "", err
	}

	return unmarshalHover(payload.Result.Contents)
}

func unmarshalHover(blah json.RawMessage) (string, error) {
	var target []json.RawMessage
	if err := json.Unmarshal(blah, &target); err != nil {
		return unmarshalHoverPart(blah)
	}

	var parts []string
	for _, t := range target {
		part, err := unmarshalHoverPart(t)
		if err != nil {
			return "", err
		}

		parts = append(parts, part)
	}

	return strings.Join(parts, "\n\n---\n\n"), nil
}

func unmarshalHoverPart(blah json.RawMessage) (string, error) {
	var p string
	if err := json.Unmarshal(blah, &p); err == nil {
		return strings.TrimSpace(p), nil
	}

	var payload struct {
		Kind     string `json:"kind"`
		Language string `json:"language"`
		Value    string `json:"value"`
	}
	if err := json.Unmarshal(blah, &payload); err != nil {
		return "", errors.New("unrecognized hover format")
	}

	if payload.Language != "" {
		return fmt.Sprintf("```%s\n%s\n```", payload.Language, payload.Value), nil
	}

	return strings.TrimSpace(payload.Value), nil
}
