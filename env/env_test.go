package env

import (
	"testing"
)

func TestGetEnvOr(t *testing.T) {
	t.Setenv("TEST_INT", "42")

	result := GetEnvOr("TEST_INT", 0)
	if result != 42 {
		t.Fatal("Error result must be 42")
	}

	result = GetEnvOr("NO_EXISTENT", 42)
	if result != 42 {
		t.Fatal("Error result must be 42")
	}
}
