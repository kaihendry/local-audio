package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/apex/gateway/v2"
	"github.com/apex/log"
	jsonhandler "github.com/apex/log/handlers/json"
	"github.com/apex/log/handlers/text"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Record struct {
	ID        string    `dynamodbav:"id" json:"id"`
	Created   time.Time `dynamodbav:"created,unixtime" json:"created"`
	Expires   time.Time `dynamodbav:"expires,unixtime" json:"expires"`
	Latitude  float64   `dynamodbav:"latitude" json:"latitude"`
	Longitude float64   `dynamodbav:"longitude" json:"longitude"`
	Title     string    `dynamodbav:"title" json:"title"`
	Audio     string    `dynamodbav:"audio" json:"audio"`
}

type server struct {
	router *http.ServeMux
	db     *dynamodb.Client
	store  *s3.Client
}

// Routes are defined here
func newServer(local bool) *server {
	s := &server{router: &http.ServeMux{}}

	if local {
		log.SetHandler(text.Default)
		log.Info("local mode")
		s.db = dynamoLocal()
	} else {
		log.SetHandler(jsonhandler.Default)
		log.Info("cloud mode")
		s.db = dynamoCloud()
	}

	s.store = s3Cloud()

	s.router.Handle("/", s.index())
	s.router.Handle("/add", s.add())

	return s
}

func main() {
	_, awsDetected := os.LookupEnv("AWS_LAMBDA_FUNCTION_NAME")
	log.WithField("awsDetected", awsDetected).Info("starting up")
	s := newServer(!awsDetected)

	var err error

	if awsDetected {
		err = gateway.ListenAndServe("", s.router)
	} else {
		err = http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), s.router)
	}
	log.WithError(err).Fatal("error listening")
}
