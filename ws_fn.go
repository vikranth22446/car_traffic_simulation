package main

import (
	"fmt"
	"github.com/google/uuid"
)

func identFn(id uuid.UUID) []byte {
	s := fmt.Sprintf(`[{"fn": "identify", "id": "%s"}]`, id.String())
	return []byte(s)
}

func createFn(t string, id uuid.UUID, data []byte) []byte {
	s := fmt.Sprintf(`[{"fn": "create", "kind": "%s", "id": "%s", "data": %s}]`, t, id.String(), string(data))
	return []byte(s)
}

func updateFn(id string, data []byte) []byte {
	s := fmt.Sprintf(`[{"fn": "update", "id": "%s", "data": %s}]`, id, string(data))
	return []byte(s)
}
