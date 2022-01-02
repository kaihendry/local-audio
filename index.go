package main

import (
	"context"
	"embed"
	"html/template"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/apex/log"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type Point struct {
	lat float64
	lng float64
}

//go:embed templates
var tmpl embed.FS

func (record *Record) TimeSinceCreation() int {
	// return minutes since created
	return int(time.Since(record.Created).Minutes())

}

func hsin(theta float64) float64 {
	return math.Pow(math.Sin(theta/2), 2)
}

func (record *Record) Distance(p Point) (distance float64) {
	// https://gist.github.com/cdipaolo/d3f8db3848278b49db68
	var la1, lo1, la2, lo2, r float64
	la1 = record.Latitude * math.Pi / 180
	lo1 = record.Longitude * math.Pi / 180
	la2 = p.lat * math.Pi / 180
	lo2 = p.lng * math.Pi / 180

	r = 6378100 // Earth radius in METERS

	h := hsin(la2-la1) + math.Cos(la1)*math.Cos(la2)*hsin(lo2-lo1)

	return 2 * r * math.Asin(math.Sqrt(h))
}

func (record *Record) TimeUntilExpiry() string {
	return time.Until(record.Expires).String()
}

func (s *server) index() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("list")

		t, err := template.ParseFS(tmpl, "templates/*.html")
		if err != nil {
			log.WithError(err).Fatal("Failed to parse templates")
		}

		// https://aws.github.io/aws-sdk-go-v2/docs/code-examples/dynamodb/scanitems/
		records, err := s.db.Scan(context.TODO(), &dynamodb.ScanInput{
			TableName: aws.String(os.Getenv("TABLE_NAME")),
		})
		if err != nil {
			log.WithError(err).Fatal("couldn't get records")
		}

		log.WithField("table", os.Getenv("TABLE_NAME")).Info("got records")

		var selection []Record

		err = attributevalue.UnmarshalListOfMaps(records.Items, &selection)
		if err != nil {
			log.WithError(err).Fatal("couldn't parse records")
		}

		// grab lat and lng from get params
		lat, err := strconv.ParseFloat(r.FormValue("lat"), 64)
		if err != nil {
			log.WithError(err).Warn("couldn't parse lat")
		}

		lng, err := strconv.ParseFloat(r.FormValue("lng"), 64)
		if err != nil {
			log.WithError(err).Warn("couldn't parse lng")
		}

		log.WithField("count", len(selection)).Info("parsed records")

		w.Header().Set("Content-Type", "text/html")
		err = t.ExecuteTemplate(w, "index.html", struct {
			Selection    []Record
			UserLocation Point
		}{
			selection,
			Point{lat, lng},
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.WithError(err).Fatal("Failed to execute templates")
		}
	}
}
