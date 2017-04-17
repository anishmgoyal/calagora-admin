package utils

import (
	"io/ioutil"

	"github.com/anishmgoyal/calagora-admin/constants"
	"github.com/anishmgoyal/calagora-admin/models"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// LoadAttachment attempts to fetch the body of an attachment from S3
func LoadAttachment(attachment models.Attachment) ([]byte, error) {
	svc := s3.New(session.New(), &aws.Config{
		Region: aws.String(constants.S3RegionString),
	})

	out, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(constants.S3Bucket),
		Key:    aws.String(attachment.FilePath),
	})

	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(out.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
