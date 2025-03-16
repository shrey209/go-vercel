package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func main() {
	bucketName := "s3-govercel"
	key := "test.txt"
	filePath := "test.txt"

	err := uploadFileToS3(bucketName, key, filePath)
	if err != nil {
		log.Fatalf("Failed to upload file: %v", err)
	}

	fmt.Println("File uploaded successfully!")
}

func uploadFileToS3(bucket, key, filePath string) error {
	awsRegion := "ap-south-1"

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			"",
			"",
			"",
		)),
		config.WithRegion(awsRegion),
	)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %v", err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	s3Client := s3.NewFromConfig(cfg)

	_, err = s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file to S3: %v", err)
	}

	return nil
}
