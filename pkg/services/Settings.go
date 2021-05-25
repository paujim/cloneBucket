package services

import (
	"io/ioutil"
	"os"

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

func NewSettigns(filename string) (*Settings, error) {
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
