package storage

import (
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type Storage interface {
	Put(file *os.File, filename string) error
}

type storage struct {
	client *s3.S3
	bucket string
}

type Config struct {
	Url    string
	Key    string
	Secret string
	Region string
	Bucket string
}

func NewStorage(config Config) (*storage, error) {
	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(config.Key, config.Secret, ""),
		Endpoint:         aws.String(config.Url),
		Region:           aws.String(config.Region),
		S3ForcePathStyle: aws.Bool(false),
	}

	newSession, err := session.NewSession(s3Config)
	if err != nil {
		return nil, err
	}
	return &storage{s3.New(newSession), config.Bucket}, err
}

func (s *storage) Get(filename string) (io.Reader, error) {
	object := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(filename),
	}

	response, err := s.client.GetObject(object)
	if err != nil {
		return nil, err
	}

	return response.Body, nil
}

func (s *storage) Put(file *os.File, filename string) error {
	object := s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(filename),
		Body:   file,
		ACL:    aws.String((string)(types.ObjectCannedACLPublicRead)),
	}
	_, err := s.client.PutObject(&object)
	return err
}

func (s *storage) Delete(filename string) error {
	object := s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(filename),
	}
	_, err := s.client.DeleteObject(&object)
	return err
}

func (s *storage) DeleteDirectory(dir string) error {
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(dir),
	}

	for {
		output, err := s.client.ListObjectsV2(input)
		if err != nil {
			return err
		}

		if len(output.Contents) == 0 {
			break
		}

		objectsToDelete := make([]*s3.ObjectIdentifier, 0, len(output.Contents))
		for _, obj := range output.Contents {
			objectsToDelete = append(objectsToDelete, &s3.ObjectIdentifier{Key: obj.Key})
		}

		_, err = s.client.DeleteObjects(&s3.DeleteObjectsInput{
			Bucket: aws.String(s.bucket),
			Delete: &s3.Delete{
				Objects: objectsToDelete,
				Quiet:   aws.Bool(true),
			},
		})
		if err != nil {
			return err
		}

		if !aws.BoolValue(output.IsTruncated) {
			break
		}
		input.ContinuationToken = output.NextContinuationToken
	}

	return nil
}
