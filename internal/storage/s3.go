package storage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
)

type S3Uploader struct {
	client       *s3.Client
	bucket       string
	pathPrefix   string
	logger       *slog.Logger
	presigner    *s3.PresignClient
	maxRetries   int
	retryBackoff time.Duration
}

// NewS3Uploader creates an uploader and verifies that the bucket is reachable.
func NewS3Uploader(region, bucket, pathPrefix, accessKey, secretKey string) (*S3Uploader, error) {
	uploader, err := NewS3UploaderWithoutCheck(region, bucket, pathPrefix, accessKey, secretKey)
	if err != nil {
		return nil, err
	}

	if err := uploader.TestConnection(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to connect to S3: %w", err)
	}

	return uploader, nil
}

// NewS3UploaderWithoutCheck creates an uploader without contacting S3.
// Credentials come from accessKey/secretKey when set, otherwise from the default AWS chain.
func NewS3UploaderWithoutCheck(region, bucket, pathPrefix, accessKey, secretKey string) (*S3Uploader, error) {
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(region),
	}
	if accessKey != "" {
		opts = append(opts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		))
	}

	cfg, err := config.LoadDefaultConfig(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg)
	presigner := s3.NewPresignClient(client)

	return &S3Uploader{
		client:       client,
		bucket:       bucket,
		pathPrefix:   pathPrefix,
		logger:       slog.Default(),
		presigner:    presigner,
		maxRetries:   3,
		retryBackoff: 2 * time.Second,
	}, nil
}

func (u *S3Uploader) TestConnection(ctx context.Context) error {
	_, err := u.client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(u.bucket)})
	if err != nil {
		return fmt.Errorf("S3 connection test failed: %w", err)
	}
	return nil
}

func (u *S3Uploader) UploadFile(ctx context.Context, localPath string, s3Key string) (string, error) {
	file, err := os.Open(localPath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("failed to stat file: %w", err)
	}

	// Try upload with retries
	var lastErr error
	backoff := u.retryBackoff

	for attempt := 0; attempt < u.maxRetries; attempt++ {
		if attempt > 0 {
			u.logger.Info("retrying S3 upload", "attempt", attempt+1, "max_retries", u.maxRetries)
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(backoff):
			}
			backoff *= 2
		}

		// Check if file already exists (duplicate detection)
		if exists, err := u.fileExists(ctx, s3Key); err == nil && exists {
			u.logger.Info("S3 file already exists, skipping upload", "s3_key", s3Key)
			return s3Key, nil
		}

		// Reset file pointer
		file.Seek(0, 0)

		// Use multipart upload for reliability
		uploader := manager.NewUploader(u.client)
		result, err := uploader.Upload(ctx, &s3.PutObjectInput{
			Bucket:      aws.String(u.bucket),
			Key:         aws.String(s3Key),
			Body:        file,
			ContentType: aws.String("video/mp4"),
			Metadata: map[string]string{
				"file-size": fmt.Sprintf("%d", fileInfo.Size()),
				"uploaded":  time.Now().UTC().Format(time.RFC3339),
			},
		})

		if err == nil {
			u.logger.Info("file uploaded to S3", "s3_key", s3Key, "location", result.Location)
			return s3Key, nil
		}

		// Check if error is retryable
		if !IsRetryableError(err) {
			return "", fmt.Errorf("non-retryable S3 error: %w", err)
		}

		lastErr = err
	}

	return "", fmt.Errorf("max retries exceeded: %w", lastErr)
}

func (u *S3Uploader) UploadChunk(ctx context.Context, chunkPath string, cameraID string) (string, error) {
	_, err := os.Stat(chunkPath)
	if err != nil {
		return "", fmt.Errorf("failed to stat chunk file: %w", err)
	}

	// Generate S3 key based on timestamp and camera ID
	filename := filepath.Base(chunkPath)
	s3Key := fmt.Sprintf("%s/%s/%s", u.pathPrefix, cameraID, filename)

	return u.UploadFile(ctx, chunkPath, s3Key)
}

func (u *S3Uploader) UploadScreenshot(ctx context.Context, imagePath string, cameraID string) (string, error) {
	timestamp := time.Now().UTC()
	filename := fmt.Sprintf("%d-%02d-%02d/%02d-%02d-%02d.jpg",
		timestamp.Year(), timestamp.Month(), timestamp.Day(),
		timestamp.Hour(), timestamp.Minute(), timestamp.Second())
	s3Key := fmt.Sprintf("%s/%s/screenshots/%s", u.pathPrefix, cameraID, filename)

	return u.UploadFile(ctx, imagePath, s3Key)
}

func (u *S3Uploader) GetPresignedURL(ctx context.Context, s3Key string, expirationMinutes int) (string, error) {
	presignResult, err := u.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(u.bucket),
		Key:    aws.String(s3Key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(expirationMinutes) * time.Minute
	})

	if err != nil {
		return "", fmt.Errorf("failed to create presigned URL: %w", err)
	}

	return presignResult.URL, nil
}

func (u *S3Uploader) fileExists(ctx context.Context, s3Key string) (bool, error) {
	_, err := u.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(u.bucket),
		Key:    aws.String(s3Key),
	})

	if err == nil {
		return true, nil
	}

	var ae smithy.APIError
	if errors.As(err, &ae) && ae.ErrorCode() == "NotFound" {
		return false, nil
	}

	return false, err
}

func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	var ae smithy.APIError
	if !errors.As(err, &ae) {
		return false // Not an API error, possibly network-related (likely retryable)
	}

	code := ae.ErrorCode()
	// List of retryable error codes
	retryableCodes := map[string]bool{
		"ThrottlingException":  true,
		"RequestLimitExceeded": true,
		"SlowDown":             true,
		"ServiceUnavailable":   true,
		"RequestTimeout":       true,
		"OperationAborted":     true,
	}

	return retryableCodes[code]
}
