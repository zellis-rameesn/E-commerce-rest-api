package providers

import (
	"context"
	"mime/multipart"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	appConfig "github.com/zellis-rameesn/go-ecommerce/internal/config"
)

type S3Provider struct {
	client   *s3.Client
	uploader *manager.Uploader
	bucket   string
	endpoint string
}

func NewS3Provider(cfg *appConfig.AWSConfig) *S3Provider {
	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		panic("failed to create AWS config " + err.Error())
	}

	// Configure for localstack
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.S3Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.S3Endpoint)
			o.UsePathStyle = true
		}
	})

	return &S3Provider{
		client:   client,
		uploader: manager.NewUploader(client),
		bucket:   cfg.S3Bucket,
		endpoint: cfg.S3Endpoint,
	}
}

func (p *S3Provider) UploadFile(file *multipart.FileHeader, path string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	result, err := p.uploader.Upload( // for small files, like in our case we can also use client.PutObject. The uploader.Upload can upload large files as multipart uploads concurrently.
		context.TODO(),
		&s3.PutObjectInput{
			Bucket: aws.String(p.bucket),
			Key:    aws.String(path),
			Body:   src,
		},
	)
	if err != nil {
		return "", err
	}

	return *result.Key, nil
}

func (p *S3Provider) DeleteFile(path string) error {
	_, err := p.client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(strings.TrimPrefix(path, "/")), //S3 keys should not start with "/", hence removing the / prefix
	})
	return err
}
