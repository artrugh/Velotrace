package platform

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Storage struct {
	Client        *s3.Client
	PresignClient *s3.PresignClient
	Bucket        string
}

func NewStorage() (*Storage, error) {
	endpoint := os.Getenv("STORAGE_ENDPOINT")
	accesskey := os.Getenv("STORAGE_ACCESS_KEY")
	secretkey := os.Getenv("STORAGE_SECRET_KEY")
	region := os.Getenv("STORAGE_REGION")
	bucket := os.Getenv("STORAGE_BUCKET")

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accesskey, secretkey, "",
		)))

	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true

	})

	presignClient := s3.NewPresignClient(client)

	return &Storage{
		Client:        client,
		PresignClient: presignClient,
		Bucket:        bucket,
	}, nil

}
func (s *Storage) VerifyConnection(ctx context.Context) error {
	_, err := s.Client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return fmt.Errorf("storage connection failed: %w", err)
	}
	return nil
}

func (s *Storage) GetPresignedPutURL(ctx context.Context, objectKey string) (string, error) {
	params := &s3.PutObjectInput{
		Bucket:      aws.String(s.Bucket),
		Key:         aws.String(objectKey),
		ContentType: aws.String("image/jpeg"),
	}

	request, err := s.PresignClient.PresignPutObject(ctx, params, func(opts *s3.PresignOptions) {
		opts.Expires = 15 * time.Minute
	})

	if err != nil {
		return "", fmt.Errorf("failed to presign upload url: %w", err)
	}

	return request.URL, nil
}
