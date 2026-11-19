package integration

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	backendPort      = "8080/tcp"
	backendImageRepo = "rtsprec-backend-integration"

	// S3-compatible server from Docker Hub. MinIO images are not pullable anonymously.
	s3Image     = "chrislusf/seaweedfs:latest"
	s3Port      = "8333/tcp"
	s3Alias     = "seaweed"
	s3DomainArg = "-s3.domainName=" + s3Alias

	// Credentials and names used by every test stack.
	testBucket    = "test-bucket"
	s3AccessKey   = "testaccess"
	s3SecretKey   = "testsecret"
	adminUser     = "admin"
	adminPassword = "test-password"
)

// requireDocker returns an error when no Docker daemon answers.
func requireDocker(ctx context.Context) error {
	p, err := testcontainers.NewDockerProvider()
	if err != nil {
		return err
	}
	defer p.Close()
	return p.Health(ctx)
}

// repoRoot is the directory that contains the root Dockerfile.
func repoRoot() string {
	abs, err := filepath.Abs("..")
	if err != nil {
		return ".."
	}
	return abs
}

// backendEnv returns the environment the backend needs. The S3 server is reached
// through the Docker network alias "seaweed". The alias "test-bucket.seaweed" lets the
// AWS SDK's virtual-host bucket addressing resolve, together with -s3.domainName.
func backendEnv(extra map[string]string) map[string]string {
	env := map[string]string{
		"DB_PATH":                   "/home/rtsprec/rtsprec.db",
		"ENCRYPTION_KEY_PATH":       "/home/rtsprec/encryption.key",
		"HLS_OUTPUT_PATH":           "/home/rtsprec/hls",
		"HTTP_PORT":                 "8080",
		"AUTH_METHOD":               "basic",
		"AUTH_BASIC_USERNAME":       adminUser,
		"AUTH_BASIC_PASSWORD":       adminPassword,
		"S3_BUCKET":                 testBucket,
		"S3_REGION":                 "us-east-1",
		"S3_ACCESS_KEY_ID":          s3AccessKey,
		"S3_SECRET_ACCESS_KEY":      s3SecretKey,
		"AWS_ENDPOINT_URL":          "http://" + s3Alias + ":8333",
		"AWS_EC2_METADATA_DISABLED": "true",
	}
	for k, v := range extra {
		env[k] = v
	}
	return env
}

// startS3 starts SeaweedFS (S3 gateway enabled) on the network and creates the test
// bucket. It returns the container and the host endpoint used by the test client.
func startS3(ctx context.Context, nw *testcontainers.DockerNetwork) (testcontainers.Container, string, error) {
	c, err := testcontainers.Run(ctx, s3Image,
		testcontainers.WithCmd("server", "-s3", s3DomainArg, "-dir=/data"),
		testcontainers.WithExposedPorts(s3Port),
		network.WithNetwork([]string{s3Alias, testBucket + "." + s3Alias}, nw),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort(s3Port).WithStartupTimeout(2*time.Minute),
		),
	)
	if err != nil {
		return nil, "", fmt.Errorf("start s3: %w", err)
	}

	endpoint, err := c.PortEndpoint(ctx, s3Port, "http")
	if err != nil {
		return c, "", fmt.Errorf("s3 endpoint: %w", err)
	}

	client := s3.NewFromConfig(aws.Config{
		Region:      "us-east-1",
		Credentials: credentials.NewStaticCredentialsProvider(s3AccessKey, s3SecretKey, ""),
	}, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})
	if _, err := client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(testBucket)}); err != nil {
		return c, "", fmt.Errorf("create bucket: %w", err)
	}

	return c, endpoint, nil
}

// startBackend builds the backend image from the root Dockerfile and starts it on
// the network. It returns once GET /api/cameras answers 401, which means the server
// is listening and authentication is enforced.
func startBackend(ctx context.Context, nw *testcontainers.DockerNetwork, extraEnv map[string]string) (testcontainers.Container, error) {
	c, err := testcontainers.Run(ctx, "",
		testcontainers.WithDockerfile(testcontainers.FromDockerfile{
			Context:    repoRoot(),
			Dockerfile: "Dockerfile",
			Repo:       backendImageRepo,
			Tag:        "integration",
			KeepImage:  true,
		}),
		testcontainers.WithEnv(backendEnv(extraEnv)),
		testcontainers.WithExposedPorts(backendPort),
		network.WithNetwork(nil, nw),
		testcontainers.WithWaitStrategy(
			wait.ForHTTP("/api/cameras").
				WithPort(backendPort).
				WithStartupTimeout(2*time.Minute).
				WithStatusCodeMatcher(func(status int) bool { return status == http.StatusUnauthorized }),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("start backend: %w", err)
	}
	return c, nil
}
