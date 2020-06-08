package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/s3"
	"gopkg.in/yaml.v2"
)

type BucketInfo struct {
	BucketName *string `yaml:"bucket"`
	AWSRegion  *string `yaml:"region"`
	AWSProfile string  `yaml:"profile"`
}

type Settings struct {
	Source      BucketInfo `yaml:"source"`
	Destination BucketInfo `yaml:"destination"`
}

type Cloner struct {
	dstIAMClient IAM
	srcS3Client  S3
	dstS3Client  S3
	srcBucket    *string
	dstBucket    *string
}

type IAM interface {
	GetUser(input *iam.GetUserInput) (*iam.GetUserOutput, error)
}

type S3 interface {
	PutBucketPolicy(*s3.PutBucketPolicyInput) (*s3.PutBucketPolicyOutput, error)
	DeleteBucketPolicy(*s3.DeleteBucketPolicyInput) (*s3.DeleteBucketPolicyOutput, error)
	ListObjectsV2(*s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error)
	CopyObject(*s3.CopyObjectInput) (*s3.CopyObjectOutput, error)
}

func CreateSettigns(filename string) (*Settings, error) {
	yamlFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer yamlFile.Close()

	byteValue, err := ioutil.ReadAll(yamlFile)
	if err != nil {
		return nil, err
	}
	var s Settings
	err = yaml.Unmarshal(byteValue, &s)
	if err != nil {
		return nil, err
	}
	return &s, err
}

func CreateCloner(sourceS3, destinationS3 S3, destinationIAM IAM, sourceBucket, destinationBucket *string) *Cloner {
	return &Cloner{
		dstIAMClient: destinationIAM,
		srcS3Client:  sourceS3,
		dstS3Client:  destinationS3,
		srcBucket:    sourceBucket,
		dstBucket:    destinationBucket,
	}
}

func (c *Cloner) updateSourceBucketPolicy() error {

	resp, err := c.dstIAMClient.GetUser(&iam.GetUserInput{})
	if err != nil {
		return err
	}
	if resp.User.Arn == nil {
		return errors.New("Invalid iam/user")
	}
	jsonPolicy := fmt.Sprintf("{\"Version\": \"2012-10-17\", \"Statement\": [{ \"Effect\": \"Allow\",\"Principal\": {\"AWS\": \"%s\"}, \"Action\": [ \"s3:Get*\",\"s3:List*\"], \"Resource\": [\"arn:aws:s3:::%s\", \"arn:aws:s3:::%s/*\" ] } ]}", *resp.User.Arn, *c.srcBucket, *c.srcBucket)
	log.Println(jsonPolicy)
	input := &s3.PutBucketPolicyInput{
		Bucket: c.srcBucket,
		Policy: aws.String(jsonPolicy),
	}
	_, err = c.srcS3Client.PutBucketPolicy(input)
	if err != nil {
		return err
	}
	return nil
}

func (c *Cloner) deleteSourceBucketPolicy() {
	input := &s3.DeleteBucketPolicyInput{
		Bucket: c.srcBucket,
	}
	_, err := c.srcS3Client.DeleteBucketPolicy(input)
	if err != nil {
		log.Print(err.Error())
	}
	return
}

func (c *Cloner) copyBucket() error {
	resp, err := c.srcS3Client.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: c.srcBucket})
	if err != nil {
		return err
	}

	for _, item := range resp.Contents {
		copyKey := url.QueryEscape(*c.srcBucket + "/" + *item.Key)
		_, err := c.dstS3Client.CopyObject(&s3.CopyObjectInput{
			Bucket:     c.dstBucket,
			CopySource: aws.String(copyKey),
			Key:        item.Key,
		})
		if err != nil {
			log.Print(err.Error())
		}
	}
	return nil
}

func (c *Cloner) Clone() error {

	err := c.updateSourceBucketPolicy()
	if err != nil {
		return err
	}
	time.Sleep(500 * time.Microsecond)

	err = c.copyBucket()
	c.deleteSourceBucketPolicy()
	return err
}
