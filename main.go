package main

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

var (
	createS3Client = func(region *string, profile string) (s3iface.S3API, error) {
		sess, err := session.NewSession(&aws.Config{
			Region:      region,
			Credentials: credentials.NewSharedCredentials("", profile),
		})
		if err != nil {
			return nil, err
		}
		return s3.New(sess), nil
	}
	createIAMClient = func(profile string) (iamiface.IAMAPI, error) {
		sess, err := session.NewSession(&aws.Config{
			Credentials: credentials.NewSharedCredentials("", profile),
		})
		if err != nil {
			return nil, err
		}
		return iam.New(sess), nil
	}
)

func main() {

	settings, err := CreateSettigns("settings.yaml")
	if err != nil {
		log.Printf("ERROR: %s\n", err.Error())
		return
	}
	sourceS3, err := createS3Client(settings.Source.AWSRegion, settings.Source.AWSProfile)
	if err != nil {
		log.Printf("ERROR: %s\n", err.Error())
		return
	}
	destinationS3, err := createS3Client(settings.Destination.AWSRegion, settings.Destination.AWSProfile)
	if err != nil {
		log.Printf("ERROR: %s\n", err.Error())
		return
	}
	destinationIAM, err := createIAMClient(settings.Destination.AWSProfile)

	err = CreateCloner(sourceS3, destinationS3, destinationIAM, settings.Source.BucketName, settings.Destination.BucketName).Clone()
	if err != nil {
		log.Printf("ERROR: %s\n", err.Error())
		return
	}
}
