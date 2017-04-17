package services

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/anishmgoyal/calagora-admin/constants"
	"github.com/anishmgoyal/calagora-admin/models"
	"github.com/anishmgoyal/calagora-admin/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// DownloadEmailForUser attempts to download a user's emails from the
// S3 bucket containing the emails
func DownloadEmailForUser(username string, event chan string) {
	markDownloadFinished := func() { event <- username }
	defer markDownloadFinished()

	svc := s3.New(session.New(), &aws.Config{
		Region: aws.String(constants.S3RegionString),
	})

	response, err := svc.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(constants.S3Bucket),
		Prefix: aws.String(username),
	})

	if err != nil {
		return
	}

	for _, objectInfo := range response.Contents {
		object, err := svc.GetObject(&s3.GetObjectInput{
			Bucket: aws.String(constants.S3Bucket),
			Key:    objectInfo.Key,
		})
		if err != nil {
			continue
		}
		body, err := ioutil.ReadAll(object.Body)
		if err != nil {
			continue
		}
		email := utils.ParseEmail(string(body))
		if saveEmail(email, svc) == nil {
			// We successfully downloaded the email... remove it from S3
			svc.DeleteObject(&s3.DeleteObjectInput{
				Bucket: aws.String(constants.S3Bucket),
				Key:    objectInfo.Key,
			})
		}
	}
}

func saveEmail(email *models.Email, svc *s3.S3) error {
	err := email.Create(Base.DB)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	for _, attachment := range email.Attachments {
		fileName := "attachments/" + strconv.Itoa(email.ID) + "_attachment_" +
			strconv.Itoa(attachment.ID)
		_, err := svc.PutObject(&s3.PutObjectInput{
			Bucket:      aws.String(constants.S3Bucket),
			Key:         aws.String(fileName),
			ContentType: aws.String(attachment.ContentType),
			Body:        bytes.NewReader(attachment.RawData),
		})
		if err == nil {
			attachment.FilePath = fileName
			err = attachment.Save(Base.DB)
			if err != nil {
				fmt.Println(err.Error())
			}
		}
	}
	return nil
}
