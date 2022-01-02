package main

import (
	"context"

	"github.com/apex/log"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func s3Cloud() *s3.Client {
	log.Info("s3 config")
	cfg, err := config.LoadDefaultConfig(context.TODO(), func(o *config.LoadOptions) error {
		return nil
	})
	if err != nil {
		log.WithError(err).Fatal("failed to load config")
	}
	return s3.NewFromConfig(cfg)
}
