package integration

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
)

// stack is one S3 container and one backend container on a shared network.
type stack struct {
	nw      *testcontainers.DockerNetwork
	s3      testcontainers.Container
	backend testcontainers.Container
	baseURL string
}

// shared is the stack used by the HTTP scenarios. It is started once per package run.
var shared *stack

func TestMain(m *testing.M) {
	ctx := context.Background()

	if err := requireDocker(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "integration tests need a running Docker daemon: %v\n", err)
		os.Exit(1)
	}

	s, err := startStack(ctx, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start test stack: %v\n", err)
		os.Exit(1)
	}
	shared = s

	code := m.Run()
	s.terminate(ctx)
	os.Exit(code)
}

// startStack starts the S3 server and the backend. extraEnv overrides backend variables.
func startStack(ctx context.Context, extraEnv map[string]string) (*stack, error) {
	nw, err := network.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("create network: %w", err)
	}

	s3C, _, err := startS3(ctx, nw)
	if err != nil {
		return &stack{nw: nw, s3: s3C}, err
	}

	backend, err := startBackend(ctx, nw, extraEnv)
	if err != nil {
		return &stack{nw: nw, s3: s3C}, err
	}

	baseURL, err := backend.PortEndpoint(ctx, backendPort, "http")
	if err != nil {
		return &stack{nw: nw, s3: s3C, backend: backend}, fmt.Errorf("backend endpoint: %w", err)
	}

	return &stack{nw: nw, s3: s3C, backend: backend, baseURL: baseURL}, nil
}

// terminate removes the containers and the network. Errors are logged, not fatal,
// so one failed cleanup does not hide test results.
func (s *stack) terminate(ctx context.Context) {
	for _, c := range []testcontainers.Container{s.backend, s.s3} {
		if c == nil {
			continue
		}
		if err := c.Terminate(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "terminate container: %v\n", err)
		}
	}
	if s.nw != nil {
		if err := s.nw.Remove(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "remove network: %v\n", err)
		}
	}
}
