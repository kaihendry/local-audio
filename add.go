package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/apex/log"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func (s *server) add() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			log.Info("showing upload form")
			t, err := template.ParseFS(tmpl, "templates/*.html")
			if err != nil {
				log.WithError(err).Fatal("Failed to parse templates")
			}

			w.Header().Set("Content-Type", "text/html")
			err = t.ExecuteTemplate(w, "add.html", struct{}{})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				log.WithError(err).Fatal("Failed to execute templates")
			}
			return
		}

		log.Info("processing upload post")

		// parse body to a record
		var rec Record

		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		log.Debug("parsed form")

		audioFile, header, err := r.FormFile("audio")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer audioFile.Close()

		// parse title from form
		rec.Title = r.Form.Get("title")

		// parse longitude and latitude from form
		rec.Longitude, err = strconv.ParseFloat(r.Form.Get("longitude"), 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		rec.Latitude, err = strconv.ParseFloat(r.Form.Get("latitude"), 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		rec.ID = r.RemoteAddr
		rec.Created = time.Now()
		rec.Expires = rec.Created.Add(time.Hour)

		// https://aws.github.io/aws-sdk-go-v2/docs/sdk-utilities/s3/
		// Upload the audio file to S3 client s.store and get the url
		uploader := manager.NewUploader(s.store)
		result, err := uploader.Upload(context.TODO(), &s3.PutObjectInput{
			Bucket: aws.String(os.Getenv("BUCKET_NAME")),
			// id / yyyy-mm-dd / name-of-file
			Key:  aws.String(fmt.Sprintf("%s/%s/%s", rec.ID, rec.Created.Format("2006-01-02"), header.Filename)),
			ACL:  types.ObjectCannedACLPublicRead,
			Body: audioFile,
			// set audio/x-m4a as the content type
			ContentType: aws.String("audio/x-m4a"),
		})
		if err != nil {
			log.WithError(err).Error("Failed to upload audio file")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		rec.Audio = result.Location

		log.WithFields(log.Fields{
			"record":   rec,
			"s3result": result,
		}).Info("uploaded audio file")

		av, err := attributevalue.MarshalMap(rec)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = s.db.PutItem(context.TODO(), &dynamodb.PutItemInput{
			Item:      av,
			TableName: aws.String(os.Getenv("TABLE_NAME")),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusFound)

	}
}
