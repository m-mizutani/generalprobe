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

type GetLambdaLogsCallback func(logs CloudWatchLog) bool

type getLambdaLogs struct {
	target     Target
	filter     string
	queryLimit uint
	interval   uint
	callback   GetLambdaLogsCallback
	baseScene
}

type CloudWatchLog string

func (x CloudWatchLog) Bind(data interface{}) {
	if err := json.Unmarshal([]byte(x), data); err != nil {
		log.Fatalf("Fail to unmarshal CloudWatchLog: %s", x)
	}
}
func (x CloudWatchLog) Contains(key string) bool {
	return strings.Index(string(x), key) >= 0
}

func (x *Generalprobe) GetLambdaLogs(target Target, callback GetLambdaLogsCallback) *getLambdaLogs {
	scene := getLambdaLogs{
		target:     target,
		callback:   callback,
		queryLimit: 20,
		interval:   3,
	}

	return &scene
}

func (x *getLambdaLogs) Filter(filter string) *getLambdaLogs {
	x.filter = filter
	return x
}

func (x *getLambdaLogs) QueryLimit(queryLimit uint) *getLambdaLogs {
	x.queryLimit = queryLimit
	return x
}

func (x *getLambdaLogs) Interval(interval uint) *getLambdaLogs {
	x.interval = interval
	return x
}

func (x *getLambdaLogs) play() error {
	lambdaName := x.target.name()
	if lambdaName == "" {
		logger.Fatal(fmt.Printf("No such lambda function: %s", x.target))
	}

	client := cloudwatchlogs.New(x.awsSession())
	nextToken := ""
	now := time.Now().UTC()

	for n := uint(0); n <= x.queryLimit; n++ {
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
