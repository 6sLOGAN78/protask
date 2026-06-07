package aws

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"
	"github.com/6sLOGAN78/go-protask/internal/server"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Client struct {
	server *server.Server
	client *s3.Client
}

func NewS3Client(
	server *server.Server,
	cfg aws.Config,
) *S3Client {
	return &S3Client{
		server: server,
		client: s3.NewFromConfig(cfg),
	}
}

func (s *S3Client) UploadFile(
	ctx context.Context,
	bucket string,
	fileName string,
	file io.Reader,
) (string, error) {

	fileKey := fmt.Sprintf(
		"%d_%s",
		time.Now().UnixNano(),
		filepath.Base(fileName),
	)

	contentType := "application/octet-stream"

	if ext := filepath.Ext(fileName); ext != "" {
		switch ext {
		case ".png":
			contentType = "image/png"
		case ".jpg", ".jpeg":
			contentType = "image/jpeg"
		case ".pdf":
			contentType = "application/pdf"
		}
	}

	_, err := s.client.PutObject(
		ctx,
		&s3.PutObjectInput{
			Bucket:      aws.String(bucket),
			Key:         aws.String(fileKey),
			Body:        file,
			ContentType: aws.String(contentType),
		},
	)

	if err != nil {
		return "", fmt.Errorf(
			"failed to upload file: %w",
			err,
		)
	}

	return fileKey, nil
}

func (s *S3Client) CreatePresignedUrl(
	ctx context.Context,
	bucket string,
	objectKey string,
) (string, error) {

	presignClient := s3.NewPresignClient(s.client)

	url, err := presignClient.PresignGetObject(
		ctx,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(objectKey),
		},
		s3.WithPresignExpires(
			time.Hour,
		),
	)

	if err != nil {
		return "", err
	}

	return url.URL, nil
}

func (s *S3Client) DeleteObject(
	ctx context.Context,
	bucket string,
	key string,
) error {

	_, err := s.client.DeleteObject(
		ctx,
		&s3.DeleteObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		},
	)

	if err != nil {
		return fmt.Errorf(
			"failed deleting object %s: %w",
			key,
			err,
		)
	}

	return nil
}