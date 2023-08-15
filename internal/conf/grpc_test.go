package conf

import (
	"context"
	"testing"
)

func TestIsGRPCEnabled(t *testing.T) {
	t.Setenv(envGRPCEnabled, "true")
	if !IsGRPCEnabled(context.Background()) {
		t.Fatal("expected grpc to be enabled")
	}

	t.Setenv(envGRPCEnabled, "false")
	if IsGRPCEnabled(context.Background()) {
		t.Fatal("expected grpc to not be enabled")
	}

	t.Setenv(envGRPCEnabled, "")
	if IsGRPCEnabled(context.Background()) {
		t.Fatal("expected grpc to not be enabled")
	}
}
