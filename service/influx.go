package service

import (
	"log"
	"net/url"
	"os"

	influx "github.com/influxdata/influxdb1-client"
)

var (
	INFLUX_URL  = "127.0.0.1:8086"
	INFLUX_USER = "dbu1"
	INFLUX_PASS = ""
)

func init() {
	urlString := os.Getenv("INFLUX_URL")
	if urlString != "" {
		INFLUX_URL = urlString
	}
}

func influxClient() (*influx.Client, error) {
	host, err := url.Parse("http://" + INFLUX_URL)
	if err != nil {
		return nil, err
	}

	conf := influx.Config{
		URL:      *host,
		Username: INFLUX_USER,
		Password: INFLUX_PASS,
		// URL: os.Getenv("INFLUX_URL"),
		// Username: os.Getenv("INFLUX_USER"),
		// Password: os.Getenv("INFLUX_PWD"),
	}
	con, err := influx.NewClient(conf)
	if err != nil {
		return nil, err
	}

	return con, nil
}

func influxSetRetentionPolicy(db *influx.Client) error {
	q := influx.Query{
		Command: `CREATE RETENTION POLICY "6months" ON "report" DURATION 8w REPLICATION 1`,
	}
	//SHOW RETENTION POLICIES ON report
	if response, err := db.Query(q); err == nil && response.Error() == nil {
		log.Println(response.Results)
	} else {
		return err
	}
	return nil
}
