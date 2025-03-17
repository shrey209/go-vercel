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

func uploadFile(client *s3.Client, bucket, key, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file to S3: %v", err)
	}

	fmt.Println("Uploaded:", key)
	return nil
}

func uploadDirectory(client *s3.Client, bucket, baseDir string) error {
	return filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relPath, _ := filepath.Rel(baseDir, path)
			s3Path := "vite-project/dist/" + relPath
			return uploadFile(client, bucket, s3Path, path)
		}
		return nil
	})
}

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	bucket := "s3-govercel"
	directory := "code-storage/t1/vite-project/dist"

	awsAccessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	awsSecretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	awsRegion := os.Getenv("AWS_REGION")

	if awsAccessKey == "" || awsSecretKey == "" || awsRegion == "" {
		log.Fatal("AWS credentials not found in environment variables")
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			awsAccessKey, awsSecretKey, "",
		)),
		config.WithRegion(awsRegion),
	)
	if err != nil {
		log.Fatal("Failed to load AWS config:", err)
	}

	client := s3.NewFromConfig(cfg)

	err = uploadDirectory(client, bucket, directory)
	if err != nil {
		log.Fatal("Upload failed:", err)
	}
}
