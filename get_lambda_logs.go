package generalprobe

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// GetLambdaLogsCallback is a callback type of GetLambdaLogs
type GetLambdaLogsCallback func(logs CloudWatchLog) bool

// GetLambdaLogsScene is a scene of waiting AWS Lambda logs
type GetLambdaLogsScene struct {
	target   Target
	filter   string
	callback GetLambdaLogsCallback
	pollingScene
}

// CloudWatchLog come from message part of CloudWatch Logs Events.
// The type provides utility methods for tests.
type CloudWatchLog string

// Bind marshal json to structure. If error, test will exit by log.Fatalf
func (x CloudWatchLog) Bind(data interface{}) {
	if err := json.Unmarshal([]byte(x), data); err != nil {
		log.Fatalf("Fail to unmarshal CloudWatchLog: %s", x)
	}
}

// Contains search string in the log message
func (x CloudWatchLog) Contains(key string) bool {
	return strings.Index(string(x), key) >= 0
}

// GetLambdaLogs creates a new scene to wait AWS Lambda output from CloudWatchLogs
func GetLambdaLogs(target Target, callback GetLambdaLogsCallback) *GetLambdaLogsScene {
	scene := GetLambdaLogsScene{
		target:   target,
		callback: callback,
		pollingScene: pollingScene{
			limit:    20,
			interval: 3,
		},
	}

	return &scene
}

// Filter sets filtering keyword to search CloudWatch Logs.
// Default is empty. The filter keyword will be quote automatically when querying.
func (x *GetLambdaLogsScene) Filter(filter string) *GetLambdaLogsScene {
	x.filter = filter
	return x
}

// Strings return text explanation of the scene
func (x *GetLambdaLogsScene) string() string {
	return fmt.Sprintf("Reading Lambda Logs of %s", x.target.arn(x.gp))
}

func (x *GetLambdaLogsScene) play() error {
	lambdaName := x.target.name(x.gp)
	if lambdaName == "" {
		logger.Fatal(fmt.Printf("No such lambda function: %s", x.target))
	}

	client := cloudwatchlogs.New(x.awsSession())
	nextToken := ""
	now := time.Now().UTC()

	for n := 0; n <= x.limit; n++ {
		input := cloudwatchlogs.FilterLogEventsInput{
			LogGroupName: aws.String(fmt.Sprintf("/aws/lambda/%s", lambdaName)),
			StartTime:    toMilliSec(x.startTime().Add(time.Minute * -1)),
			EndTime:      toMilliSec(now.Add(time.Minute * 1)),
		}
		if x.filter != "" {
			input.FilterPattern = aws.String(fmt.Sprintf("\"%s\"", x.filter))
		}
		if nextToken != "" {
			input.NextToken = aws.String(nextToken)
		}

		logger.WithField("input", input).Debug("Call FilterLogEvents")
		resp, err := client.FilterLogEvents(&input)
		logger.WithFields(logrus.Fields{
			"resp":  resp,
			"input": input,
			"start": *input.StartTime,
		}).Trace("Filtered log events")

		if nil != err {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case cloudwatchlogs.ErrCodeResourceNotFoundException:
					continue
				}
			}

			logger.Fatal("Can not access to ClodwatchLogs", err)
		}

		for _, event := range resp.Events {
			if event.Message != nil {
				if x.callback(CloudWatchLog(*event.Message)) {
					return nil
				}
			}
		}

		if resp.NextToken == nil {
			time.Sleep(time.Second * time.Duration(x.interval))
		} else {
			nextToken = *resp.NextToken
		}
	}

	if !x.callback(CloudWatchLog("")) {
		return errors.New("No expected logs from CloudWatch Logs")
	}

	return nil
}
