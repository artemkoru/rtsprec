package app

import (
	"fmt"

	"github.com/artemkoru/rtsprec/internal/storage"
)

// InitS3 initializes the S3 storage client.
func InitS3(region, bucket, pathPrefix, accessKey, secretKey string) (*storage.S3Uploader, error) {
	uploader, err := storage.NewS3Uploader(region, bucket, pathPrefix, accessKey, secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize S3 uploader: %w", err)
	}

	return uploader, nil
}
