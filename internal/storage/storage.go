package storage

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type Storage interface {
	Get(filename string) (io.Reader, error)
	Put(file *os.File, remoteName string) error
	Delete(filename string) error
	DeleteDirectory(dir string) error
}

type storage struct {
	client *s3.Client
	bucket string
}

type Config struct {
	Url          string
	Region       string
	Bucket       string
	BucketKey    string
	BucketSecret string
}

func NewStorage(cfg Config) *storage {
	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.BucketKey,
			cfg.BucketSecret,
			"",
		)),
		config.WithEndpointResolver(aws.EndpointResolverFunc(
			func(service, region string) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:           cfg.Url,
					SigningRegion: cfg.Region,
				}, nil
			},
		)),
	)
	if err != nil {
		log.Fatal(err)
	}
	client := s3.NewFromConfig(awsCfg)
	return &storage{client: client, bucket: cfg.Bucket}
}

func (s *storage) Get(remoteName string) (io.Reader, error) {
	object := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(remoteName),
	}

	response, err := s.client.GetObject(context.TODO(), object)
	if err != nil {
		return nil, err
	}

	return response.Body, nil
}

func (s *storage) Put(file *os.File, remoteName string) error {
	object := s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(remoteName),
		Body:   file,
		ACL:    types.ObjectCannedACLPublicRead,
	}

	_, err := s.client.PutObject(context.TODO(), &object)
	return err
}

func (s *storage) Delete(remoteName string) error {
	object := s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(remoteName),
	}
	_, err := s.client.DeleteObject(context.TODO(), &object)
	return err
}

func (s *storage) DeleteDirectory(dir string) error {
	input := &s3.ListObjectsV2Input{
		Bucket: &s.bucket,
		Prefix: &dir,
	}

	for {
		output, err := s.client.ListObjectsV2(context.TODO(), input)
		if err != nil {
			return err
		}

		if len(output.Contents) == 0 {
			break
		}

		objectsToDelete := make([]types.ObjectIdentifier, 0, len(output.Contents))
		for _, obj := range output.Contents {
			objectsToDelete = append(objectsToDelete, types.ObjectIdentifier{Key: obj.Key})
		}

		_, err = s.client.DeleteObjects(context.TODO(), &s3.DeleteObjectsInput{
			Bucket: &s.bucket,
			Delete: &types.Delete{
				Objects: objectsToDelete,
				Quiet:   aws.Bool(true),
			},
		})
		if err != nil {
			return err
		}

		if output.IsTruncated == nil || !*output.IsTruncated {
			break
		}
		input.ContinuationToken = output.NextContinuationToken
	}

	return nil
}
