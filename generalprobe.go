package generalprobe

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"

	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

func init() {
	level := os.Getenv("GENERALPROBE_LOG_LEVEL")

	switch {
	case strings.EqualFold(level, "debug"):
		logger.SetLevel(logrus.DebugLevel)
	case strings.EqualFold(level, "info"):
		logger.SetLevel(logrus.InfoLevel)
	case strings.EqualFold(level, "warn"):
		logger.SetLevel(logrus.WarnLevel)
	case strings.EqualFold(level, "error"):
		logger.SetLevel(logrus.ErrorLevel)
	default:
		logger.SetLevel(logrus.WarnLevel)
	}
}

// SetLoggerDebugLevel changes logging level to Debug
func SetLoggerDebugLevel() { logger.SetLevel(logrus.DebugLevel) }

// SetLoggerInfoLevel changes logging level to Info
func SetLoggerInfoLevel() { logger.SetLevel(logrus.InfoLevel) }

// SetLoggerWarnLevel changes logging level to Warn
func SetLoggerWarnLevel() { logger.SetLevel(logrus.WarnLevel) }

// Generalprobe is a main structure of the framework
type Generalprobe struct {
	awsRegion  string
	awsSession *session.Session
	awsAccount string
	stackName  string
	stackArn   string
	scenes     []Scene
	resources  []*cloudformation.StackResource
	done       bool

	StartTime time.Time
}

// New is constructor of Generalprobe structure.
func New(awsRegion, stackName string) Generalprobe {
	gp := Generalprobe{
		awsRegion: awsRegion,
		stackName: stackName,
		done:      false,
		StartTime: time.Now().UTC(),
	}

	gp.awsSession = session.Must(session.NewSession(&aws.Config{
		Region: aws.String(awsRegion),
	}))
	client := cloudformation.New(gp.awsSession)

	resp, err := client.DescribeStackResources(&cloudformation.DescribeStackResourcesInput{
		StackName: aws.String(stackName),
	})
	// pp.Println(resp)
	if err != nil {
		logger.Fatal("Fail to get CloudFormation Stack resources: ", err)
	}

	gp.resources = resp.StackResources

	stackResp, err := client.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})

	if err != nil {
		logger.Fatal("Fail to get detail of CloudFormation Stack instance", err, stackName)
	}

	for _, stack := range stackResp.Stacks {
		if *stack.StackName == stackName {
			gp.stackArn = *stack.StackId
			sec := strings.Split(gp.stackArn, ":")
			gp.awsAccount = sec[4]
		}
	}

	return gp
}

// LookupID looks up PhysicalID from resource list of the CFn stack.
func (x *Generalprobe) LookupID(logicalID string) string {
	for _, resource := range x.resources {
		if resource.LogicalResourceId != nil && *resource.LogicalResourceId == logicalID {
			return *resource.PhysicalResourceId
		}
	}

	return ""
}

// LookupType looks up ResourceType
func (x *Generalprobe) LookupType(logicalID string) string {
	for _, resource := range x.resources {
		if resource.LogicalResourceId != nil && *resource.LogicalResourceId == logicalID {
			return *resource.ResourceType
		}
	}

	return ""
}

func toMilliSec(t time.Time) *int64 {
	var u int64
	u = (t.Unix() * 1000)
	return &u
}

type SearchLambdaLogsArgs struct {
	LambdaTarget Target
	Filter       string
	QueryLimit   uint
	Interval     uint
}

// SearchLambdaLogs sends query to ClodWatchLogs and retrieve logs output by Lambda
func (x *Generalprobe) SearchLambdaLogs(args SearchLambdaLogsArgs) []string {
	const defaultQueryLimit = 20
	const defaultInterval = 3

	if args.QueryLimit == 0 {
		args.QueryLimit = defaultQueryLimit
	}
	if args.Interval == 0 {
		args.Interval = defaultInterval
	}

	var result []string
	lambdaName := args.LambdaTarget.name()
	if lambdaName == "" {
		logger.Fatal(fmt.Printf("No such lambda function: %s", args.LambdaTarget))
	}

	client := cloudwatchlogs.New(x.awsSession)
	nextToken := ""
	now := time.Now().UTC()

	for n := uint(0); n <= args.QueryLimit; n++ {
		input := cloudwatchlogs.FilterLogEventsInput{
			LogGroupName: aws.String(fmt.Sprintf("/aws/lambda/%s", lambdaName)),
			StartTime:    toMilliSec(x.StartTime.Add(time.Minute * -1)),
			EndTime:      toMilliSec(now.Add(time.Minute * 1)),
		}
		if args.Filter != "" {
			input.FilterPattern = aws.String(fmt.Sprintf("\"%s\"", args.Filter))
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
				result = append(result, *event.Message)
			}
		}

		if resp.NextToken == nil {
			if len(result) == 0 {
				time.Sleep(time.Second * time.Duration(args.Interval))
				continue
			}
			break
		}
		nextToken = *resp.NextToken
	}

	return result
}

// AddScenes appends Scene set to Generalprobe instance.
func (x *Generalprobe) AddScenes(newScenes []Scene) {
	for _, scene := range newScenes {
		scene.setGeneralprobe(x)
		x.scenes = append(x.scenes, scene)
	}
}

// Run invokes test according to appended Scenes.
func (x *Generalprobe) Run() error {
	for idx, scene := range x.scenes {
		logger.Infof("Step (%d/%d): %s\n", idx+1, len(x.scenes), reflect.TypeOf(scene))

		if err := scene.play(); err != nil {
			logger.WithFields(logrus.Fields{
				"sceneType": reflect.TypeOf(scene),
				"sceneNo":   idx,
				"scene":     scene,
				"error":     err,
			}).Error("Failed Generalprobe")
			return err
		}
	}

	return nil
}

func (x *Generalprobe) LogicalID(logicalID string) *LogicalIDTarget {
	r := newLogicalID(logicalID)
	r.setGeneralprobe(x)
	return r
}

func (x *Generalprobe) Arn(logicalID string) *ArnTarget {
	r := newArn(logicalID)
	r.setGeneralprobe(x)
	return r
}
