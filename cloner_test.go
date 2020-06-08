package main

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockS3 struct {
	mock.Mock
}

func (m *MockS3) PutBucketPolicy(input *s3.PutBucketPolicyInput) (*s3.PutBucketPolicyOutput, error) {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*s3.PutBucketPolicyOutput), args.Error(1)
}

func (m *MockS3) DeleteBucketPolicy(input *s3.DeleteBucketPolicyInput) (*s3.DeleteBucketPolicyOutput, error) {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*s3.DeleteBucketPolicyOutput), args.Error(1)
}

func (m *MockS3) ListObjectsV2(input *s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error) {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*s3.ListObjectsV2Output), args.Error(1)
}

func (m *MockS3) CopyObject(input *s3.CopyObjectInput) (*s3.CopyObjectOutput, error) {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*s3.CopyObjectOutput), args.Error(1)
}

type MockIAM struct {
	mock.Mock
}

func (m *MockIAM) GetUser(input *iam.GetUserInput) (*iam.GetUserOutput, error) {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*iam.GetUserOutput), args.Error(1)
}

func TestCreateSettigns(t *testing.T) {
	t.Run("Valid yaml file", func(t *testing.T) {
		assert := assert.New(t)
		s, err := CreateSettigns("testing1.yaml")
		assert.NoError(err)
		assert.NotNil(s)
		assert.NotNil(s.Source)
		assert.NotNil(s.Destination)
		assert.Equal(s.Source.BucketName, aws.String("srcBucket"))
		assert.Equal(s.Destination.BucketName, aws.String("dstBucket"))
		assert.Equal(s.Source.AWSProfile, "srcProfile")
		assert.Equal(s.Destination.AWSProfile, "dstProfile")
	})
	t.Run("Valid yaml file with out destination bucket", func(t *testing.T) {
		assert := assert.New(t)
		s, err := CreateSettigns("testing2.yaml")
		assert.NoError(err)
		assert.NotNil(s)
		assert.NotNil(s.Source)
		assert.NotNil(s.Destination)
		assert.Equal(s.Source.BucketName, aws.String("srcBucket"))
		assert.Nil(s.Destination.BucketName)
		assert.Equal(s.Source.AWSProfile, "srcProfile")
		assert.Equal(s.Destination.AWSProfile, "dstProfile")
	})

	t.Run("Unable to open file", func(t *testing.T) {
		assert := assert.New(t)
		s, err := CreateSettigns("bad file")
		assert.Error(err)
		assert.Nil(s)
	})
	t.Run("Bad yaml file", func(t *testing.T) {
		assert := assert.New(t)
		s, err := CreateSettigns("bad.yaml")
		assert.Error(err)
		assert.Nil(s)
	})
}

func TestCreateClone(t *testing.T) {

	t.Run("Valid clone", func(t *testing.T) {
		assert := assert.New(t)
		mockIAM := &MockIAM{}
		mockIAM.On("GetUser", mock.Anything).Return(&iam.GetUserOutput{User: &iam.User{Arn: aws.String("arn")}}, nil)
		sourceS3 := &MockS3{}
		sourceS3.On("PutBucketPolicy", mock.Anything).Return(&s3.PutBucketPolicyOutput{}, nil)
		sourceS3.On("ListObjectsV2", mock.Anything).Return(&s3.ListObjectsV2Output{Contents: []*s3.Object{{Key: aws.String("key")}}}, nil)
		sourceS3.On("DeleteBucketPolicy", mock.Anything).Return(&s3.DeleteBucketPolicyOutput{}, nil)
		destinationS3 := &MockS3{}
		destinationS3.On("CopyObject", mock.Anything).Return(&s3.CopyObjectOutput{}, nil)
		err := CreateCloner(sourceS3, destinationS3, mockIAM, aws.String("src-bucket"), aws.String("dst-bucket")).Clone()
		assert.NoError(err)
		mockIAM.AssertExpectations(t)
		sourceS3.AssertExpectations(t)
		destinationS3.AssertExpectations(t)
	})
	t.Run("Fail to update bucket policy", func(t *testing.T) {
		assert := assert.New(t)
		mockIAM := &MockIAM{}
		mockIAM.On("GetUser", mock.Anything).Return(&iam.GetUserOutput{User: &iam.User{Arn: aws.String("arn")}}, nil)
		sourceS3 := &MockS3{}
		sourceS3.On("PutBucketPolicy", mock.Anything).Return(nil, errors.New("Fail to update policy"))
		destinationS3 := &MockS3{}
		err := CreateCloner(sourceS3, destinationS3, mockIAM, aws.String("src-bucket"), aws.String("dst-bucket")).Clone()
		assert.EqualError(err, "Fail to update policy")
		mockIAM.AssertExpectations(t)
		sourceS3.AssertExpectations(t)
		destinationS3.AssertExpectations(t)
	})
	t.Run("Fail to list source bucket", func(t *testing.T) {
		assert := assert.New(t)
		mockIAM := &MockIAM{}
		mockIAM.On("GetUser", mock.Anything).Return(&iam.GetUserOutput{User: &iam.User{Arn: aws.String("arn")}}, nil)
		sourceS3 := &MockS3{}
		sourceS3.On("PutBucketPolicy", mock.Anything).Return(&s3.PutBucketPolicyOutput{}, nil)
		sourceS3.On("ListObjectsV2", mock.Anything).Return(nil, errors.New("Fail to list objects"))
		sourceS3.On("DeleteBucketPolicy", mock.Anything).Return(&s3.DeleteBucketPolicyOutput{}, nil)
		destinationS3 := &MockS3{}
		err := CreateCloner(sourceS3, destinationS3, mockIAM, aws.String("src-bucket"), aws.String("dst-bucket")).Clone()
		assert.EqualError(err, "Fail to list objects")
		mockIAM.AssertExpectations(t)
		sourceS3.AssertExpectations(t)
		destinationS3.AssertExpectations(t)
	})
	t.Run("Fail to get iam/user", func(t *testing.T) {
		assert := assert.New(t)
		mockIAM := &MockIAM{}
		mockIAM.On("GetUser", mock.Anything).Return(nil, errors.New("Fail to get iam/user"))
		sourceS3 := &MockS3{}
		destinationS3 := &MockS3{}
		err := CreateCloner(sourceS3, destinationS3, mockIAM, aws.String("src-bucket"), aws.String("dst-bucket")).Clone()
		assert.EqualError(err, "Fail to get iam/user")
		mockIAM.AssertExpectations(t)
		sourceS3.AssertExpectations(t)
		destinationS3.AssertExpectations(t)
	})
}
