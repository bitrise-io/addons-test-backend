package metrics

import (
	"fmt"
	"os"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/bitrise-io/api-utils/logging"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// buffer up to 10 commands
const (
	dogStatsDefaultAddress     = "127.0.0.1:8125"
	dogStatsDMetricsBufferSize = 10
	dogStatsDNamespace         = "bitrise"
	dogStatsDSubsystem         = "addons-test"
)

// DogStatsDInterface ...
type DogStatsDInterface interface {
	Track(t Trackable, metricName string, customTags ...string)
	Close()
}

// DogStatsDMetrics ...
type DogStatsDMetrics struct {
	client *statsd.Client
	logger *zap.Logger
}

// Taggable represents an entity that has tags or labels attached to it
type Taggable interface {
	GetTagArray() []string
}

// Trackable defines a configuration of a
// trackable piece of the execution stack
// It's used to track supervisor proccess stacks
type Trackable interface {
	Taggable

	GetProfileName() string
}

// NewDogStatsDMetrics ...
func NewDogStatsDMetrics(addr string, l *zap.Logger) *DogStatsDMetrics {
	if addr == "" {
		addr = dogStatsDefaultAddress
	}

	c, err := statsd.NewBuffered(addr, dogStatsDMetricsBufferSize)
	if err != nil {
		panic(err)
	}

	c.Namespace = fmt.Sprintf("%s.%s.", dogStatsDNamespace, dogStatsDSubsystem)
	c.Tags = append(c.Tags, fmt.Sprintf("environment:%s", os.Getenv("GO_ENV")))

	return &DogStatsDMetrics{
		client: c,
		logger: l,
	}
}

func (b *DogStatsDMetrics) createTagArray(t Taggable, tags ...string) []string {
	ret := make([]string, len(t.GetTagArray()))
	copy(ret, t.GetTagArray())
	ret = append(ret, tags...)

	return ret
}

// Track ...
func (b *DogStatsDMetrics) Track(t Trackable, metricName string, customTags ...string) {
	defer logging.Sync(b.logger)

	applicationTags := []string{fmt.Sprintf("name:%s", t.GetProfileName())}
	applicationTags = append(applicationTags, customTags...)
	tags := b.createTagArray(t, applicationTags...)

	if err := b.client.Incr(metricName, tags, 1.0); err != nil {
		b.logger.Error("DogStatsD Diagnostic backend has failed to track",
			zap.String("profile_name", t.GetProfileName()),
			zap.Any("error_details", errors.WithStack(err)),
		)
	}
}

// Close ...
func (b *DogStatsDMetrics) Close() {
	defer logging.Sync(b.logger)

	if err := b.client.Flush(); err != nil {
		b.logger.Error("DogStatsD Diagnostic backend has failed to flush its metrics",
			zap.Any("error_details", errors.WithStack(err)),
		)
	}

	if err := b.client.Close(); err != nil {
		b.logger.Error("DogStatsD Diagnostic backend has failed to close its client",
			zap.Any("error_details", errors.WithStack(err)),
		)
	}
}
