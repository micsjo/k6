package cloud

import (
	"context"
	"fmt"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/loadimpact/k6/lib"
	"github.com/loadimpact/k6/stats"
)

type Collector struct {
	referenceID string

	thresholds map[string][]string
	client     *Client
}

func New(fname string, opts lib.Options) (*Collector, error) {
	referenceID := os.Getenv("K6CLOUD_REFERENCEID")
	token := os.Getenv("K6CLOUD_TOKEN")

	thresholds := make(map[string][]string)

	for name, t := range opts.Thresholds {
		for _, threshold := range t.Thresholds {
			thresholds[name] = append(thresholds[name], threshold.Source)
		}
	}

	return &Collector{
		referenceID: referenceID,
		thresholds:  thresholds,
		client:      NewClient(token),
	}, nil
}

func (c *Collector) String() string {
	return "Load Impact"
}

func (c *Collector) Run(ctx context.Context) {
	name := os.Getenv("K6CLOUD_NAME")
	if name == "" {
		name = "k6 test"
	}

	// TODO fix this and add proper error handling
	if c.referenceID == "" {
		response := c.client.CreateTestRun(name, c.thresholds)
		if response != nil {
			c.referenceID = response.ReferenceID
		}
	}

	t := time.Now()
	<-ctx.Done()
	s := time.Now()

	log.Debug(fmt.Sprintf("http://localhost:5000/v1/metrics/%s/%d000/%d000\n", c.referenceID, t.Unix(), s.Unix()))
}

func (c *Collector) Collect(samples []stats.Sample) {

	var cloudSamples []*Sample
	for _, sample := range samples {
		sampleJSON := &Sample{
			Type:   "Point",
			Metric: sample.Metric.Name,
			Data: SampleData{
				Time:  sample.Time,
				Value: sample.Value,
				Tags:  sample.Tags,
			},
		}
		cloudSamples = append(cloudSamples, sampleJSON)
	}

	if len(cloudSamples) > 0 && c.referenceID != "" {
		c.client.PushMetric(c.referenceID, cloudSamples)
	}
}
