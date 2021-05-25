package services

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/assert"
)

func TestSettigns(t *testing.T) {
	t.Run("Valid yaml file", func(t *testing.T) {
		assert := assert.New(t)
		s, err := NewSettigns("testing1.yaml")
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
		s, err := NewSettigns("testing2.yaml")
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
		s, err := NewSettigns("bad file")
		assert.Error(err)
		assert.Nil(s)
	})
	t.Run("Bad yaml file", func(t *testing.T) {
		assert := assert.New(t)
		s, err := NewSettigns("bad.yaml")
		assert.Error(err)
		assert.Nil(s)
	})
}
