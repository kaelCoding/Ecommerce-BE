package r2

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	S3Client   *s3.Client
	BucketName string
	PublicURL  string
)

func Init() {
	accountID := os.Getenv("R2_ACCOUNT_ID")
	accessKey := os.Getenv("R2_ACCESS_KEY_ID")
	secretKey := os.Getenv("R2_SECRET_ACCESS_KEY")
	BucketName = os.Getenv("R2_BUCKET_NAME")
	PublicURL = os.Getenv("R2_PUBLIC_URL")

	if accountID == "" || accessKey == "" || secretKey == "" || BucketName == "" || PublicURL == "" {
		log.Fatal("Missing R2 environment variables (R2_ACCOUNT_ID, R2_ACCESS_KEY_ID, R2_SECRET_ACCESS_KEY, R2_BUCKET_NAME, R2_PUBLIC_URL)")
	}

	r2Endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)

	creds := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""))

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("auto"),
		config.WithCredentialsProvider(creds),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL: r2Endpoint,
				}, nil
			},
		)),
	)
	if err != nil {
		log.Fatalf("Failed to load R2 config: %v", err)
	}

	S3Client = s3.NewFromConfig(cfg)
	log.Println("Cloudflare R2 initialized successfully")
}

func UploadToR2(file io.Reader, folder string, filename string, contentType string) (string, error) {
	
    objectKey := fmt.Sprintf("%s/%s", folder, filename)

	_, err := S3Client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:      &BucketName,
		Key:         &objectKey,
		Body:        file,
		ContentType: &contentType, 
	})

	if err != nil {
		return "", fmt.Errorf("could not upload file to R2: %v", err)
	}

	fileURL := fmt.Sprintf("%s/%s", PublicURL, objectKey)
	return fileURL, nil
}