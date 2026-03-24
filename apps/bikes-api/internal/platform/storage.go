package platform

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioConfig struct {
	Endpoint    string
	AccessKey   string
	SecretKey   string
	Bucket      string
	PublicURL   string
	UseSSL      bool
	PresignHost string
	Region      string
}

func LoadMinioConfig() MinioConfig {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	if endpoint == "" {
		endpoint = "minio:9000"
	}
	presignHost := os.Getenv("MINIO_PRESIGN_HOST")
	if presignHost == "" {
		presignHost = "localhost:9000"
	}
	accessKey := os.Getenv("MINIO_ROOT_USER")
	if accessKey == "" {
		accessKey = "admin"
	}
	secretKey := os.Getenv("MINIO_ROOT_PASSWORD")
	if secretKey == "" {
		secretKey = "password123"
	}
	bucket := os.Getenv("MINIO_BUCKET")
	if bucket == "" {
		bucket = "velotrace-assets"
	}
	publicURL := os.Getenv("MINIO_PUBLIC_URL")
	if publicURL == "" {
		publicURL = "http://localhost:9000"
	}
	region := os.Getenv("MINIO_REGION")
	if region == "" {
		region = "us-east-1"
	}
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"

	return MinioConfig{
		Endpoint:    endpoint,
		AccessKey:   accessKey,
		SecretKey:   secretKey,
		Bucket:      bucket,
		PublicURL:   publicURL,
		UseSSL:      useSSL,
		PresignHost: presignHost,
		Region:      region,
	}
}

type MinioClient struct {
	Client      *minio.Client
	Bucket      string
	PublicURL   string
	PresignHost string
	AccessKey   string
	SecretKey   string
	UseSSL      bool
	Region      string
}

func InitMinio(cfg MinioConfig) (*MinioClient, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}
	if !exists {
		err = client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return &MinioClient{
		Client:      client,
		Bucket:      cfg.Bucket,
		PublicURL:   cfg.PublicURL,
		PresignHost: cfg.PresignHost,
		AccessKey:   cfg.AccessKey,
		SecretKey:   cfg.SecretKey,
		UseSSL:      cfg.UseSSL,
		Region:      cfg.Region,
	}, nil
}

func (m *MinioClient) GetPresignedPutURL(objectKey string, expiry time.Duration) (string, error) {
	// Create a separate MinIO client for presigned URLs using the host accessible outside container
	presignClient, err := minio.New(m.PresignHost, &minio.Options{
		Creds:  credentials.NewStaticV4(m.AccessKey, m.SecretKey, ""),
		Secure: m.UseSSL,
		Region: m.Region,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create presign client: %w", err)
	}
	presignedURL, err := presignClient.PresignedPutObject(context.Background(), m.Bucket, objectKey, expiry)

	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignedURL.String(), nil
}
