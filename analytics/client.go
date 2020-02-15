package analytics

import (
	"errors"
	"os"
	"time"

	"github.com/gobuffalo/uuid"
	"go.uber.org/zap"

	segment "gopkg.in/segmentio/analytics-go.v3"
)

// Interface ...
type Interface interface {
	TestReportSummaryGenerated(appSlug, buildSlug, result string, numberOfTests int, time time.Time)
	TestReportResult(appSlug, buildSlug, result, testType string, testResultID uuid.UUID, time time.Time)
	NumberOfTestReports(appSlug, buildSlug string, count int, time time.Time)
}

// Client ...
type Client struct {
	client segment.Client
	logger *zap.Logger
}

// NewClient ...
func NewClient(logger *zap.Logger) (Client, error) {
	writeKey, ok := os.LookupEnv("SEGMENT_WRITE_KEY")
	if !ok {
		return Client{}, errors.New("No value set for env SEGMENT_WRITEKEY")
	}

	return Client{
		client: segment.New(writeKey),
		logger: logger,
	}, nil
}

// TestReportSummaryGenerated ...
func (c *Client) TestReportSummaryGenerated(appSlug, buildSlug, result string, numberOfTests int, time time.Time) {
	err := c.client.Enqueue(segment.Track{
		UserId: appSlug,
		Event:  "Test report summary generated",
		Properties: segment.NewProperties().
			Set("app_slug", appSlug).
			Set("build_slug", buildSlug).
			Set("result", result).
			Set("number_of_tests", numberOfTests).
			Set("datetime", time),
	})
	if err != nil {
		c.logger.Warn("Failed to track analytics (TestReportSummaryGenerated)", zap.Error(err))
	}
}

// TestReportResult ...
func (c *Client) TestReportResult(appSlug, buildSlug, result, testType string, testResultID uuid.UUID, time time.Time) {
	err := c.client.Enqueue(segment.Track{
		UserId: appSlug,
		Event:  "Test report result",
		Properties: segment.NewProperties().
			Set("app_slug", appSlug).
			Set("build_slug", buildSlug).
			Set("result", result).
			Set("test_type", testType).
			Set("datetime", time).
			Set("test_report_id", testResultID.String()),
	})
	if err != nil {
		c.logger.Warn("Failed to track analytics (TestReportResult)", zap.Error(err))
	}
}

// NumberOfTestReports ...
func (c *Client) NumberOfTestReports(appSlug, buildSlug string, count int, time time.Time) {
	err := c.client.Enqueue(segment.Track{
		UserId: appSlug,
		Event:  "Number of test reports",
		Properties: segment.NewProperties().
			Set("app_slug", appSlug).
			Set("build_slug", buildSlug).
			Set("count", count).
			Set("datetime", time),
	})
	if err != nil {
		c.logger.Warn("Failed to track analytics (NumberOfTestReports)", zap.Error(err))
	}
}

// Close ...
func (c *Client) Close() error {
	return c.client.Close()
}
