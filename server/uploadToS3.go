package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
)

type S3Storage struct {
	client *s3.Client
	bucket string
}

func NewS3Storage() (*S3Storage, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found")
	}

	awsAccessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	awsSecretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	awsRegion := os.Getenv("AWS_REGION")
	bucket := "s3-govercel"

	if awsAccessKey == "" || awsSecretKey == "" || awsRegion == "" {
		return nil, fmt.Errorf("AWS credentials not found in environment variables")
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			awsAccessKey, awsSecretKey, "",
		)),
		config.WithRegion(awsRegion),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %v", err)
	}

	return &S3Storage{
		client: s3.NewFromConfig(cfg),
		bucket: bucket,
	}, nil
}

func (s *S3Storage) UploadFile(key, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	_, err = s.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file to S3: %v", err)
	}

	fmt.Println("Uploaded:", key)
	return nil
}

func (s *S3Storage) UploadDirectory(baseDir string) error {
	return filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relPath, _ := filepath.Rel(baseDir, path)
			s3Path := "vite-project/dist/" + relPath
			return s.UploadFile(s3Path, path)
		}
		return nil
	})
}

// func main() {
// 	storage, err := NewS3Storage()
// 	if err != nil {
// 		log.Fatal("Initialization failed:", err)
// 	}

// 	directory := "code-storage/t1/vite-project/dist"
// 	err = storage.UploadDirectory(directory)
// 	if err != nil {
// 		log.Fatal("Upload failed:", err)
// 	}
// }
