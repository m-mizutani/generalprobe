package generalprobe

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type GetLambdaLogsCallback func(logs []string) bool

type getLambdaLogs struct {
	target     Target
	filter     string
	queryLimit uint
	interval   uint
	callback   GetLambdaLogsCallback
	baseScene
}

func (x *Generalprobe) GetLambdaLogs(target Target, filter string, callback GetLambdaLogsCallback) *getLambdaLogs {
	scene := getLambdaLogs{
		target:     target,
		callback:   callback,
		queryLimit: 20,
		interval:   3,
		filter:     filter,
	}

	return &scene
}

func (x *getLambdaLogs) SetFilter(filter string) *getLambdaLogs {
	x.filter = filter
	return x
}

func (x *getLambdaLogs) SetQueryLimit(queryLimit uint) *getLambdaLogs {
	x.queryLimit = queryLimit
	return x
}

func (x *getLambdaLogs) SetInterval(interval uint) *getLambdaLogs {
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

		resp, err := client.FilterLogEvents(&input)
		logger.WithFields(logrus.Fields{
			"resp":  resp,
			"input": input,
			"start": *input.StartTime,
		}).Debug("Filtered log events")

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
				if x.callback([]string{*event.Message}) {
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

	if !x.callback([]string{}) {
		return errors.New("No expected logs from CloudWatch Logs")
	}

	return nil
}