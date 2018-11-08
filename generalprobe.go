package generalprobe

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"

	// "github.com/k0kubun/pp"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

// Generalprobe is a main structure of the framework
type Generalprobe struct {
	awsRegion  string
	awsSession *session.Session
	stackName  string
	StartTime  time.Time
	scenes     []Scene
	resources  []*cloudformation.StackResource
	done       bool
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
		log.Fatal("Fail to get CloudFormation Stack resources: ", err)
	}

	gp.resources = resp.StackResources

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

func toMilliSec(t time.Time) *int64 {
	var u int64
	u = (t.Unix() * 1000)
	return &u
}

// SearchLambdaLogs sends query to ClodWatchLogs and retrieve logs output by Lambda
func (x *Generalprobe) SearchLambdaLogs(logicalID string, filter string) []string {
	const maxRetry = 20
	const interval = 3

	var result []string
	lambdaName := x.LookupID(logicalID)
	if lambdaName == "" {
		log.Error(fmt.Printf("No such lambda function: %s", logicalID))
	}

	client := cloudwatchlogs.New(x.awsSession)
	nextToken := ""
	now := time.Now().UTC()

	for n := 0; n < maxRetry; n++ {
		input := cloudwatchlogs.FilterLogEventsInput{
			LogGroupName: aws.String(fmt.Sprintf("/aws/lambda/%s", lambdaName)),
			StartTime:    toMilliSec(x.StartTime.Add(time.Minute * -1)),
			EndTime:      toMilliSec(now.Add(time.Minute * 1)),
		}
		if filter != "" {
			input.FilterPattern = aws.String(fmt.Sprintf("\"%s\"", filter))
		}
		if nextToken != "" {
			input.NextToken = aws.String(nextToken)
		}

		resp, err := client.FilterLogEvents(&input)
		log.WithFields(log.Fields{"resp": resp, "input": input, "start": *input.StartTime}).Debug("Filtered log events")

		if err != nil {
			log.Fatal("Can not access to ClodwatchLogs", err)
		}

		for _, event := range resp.Events {
			if event.Message != nil {
				result = append(result, *event.Message)
			}
		}

		if resp.NextToken == nil {
			if len(result) == 0 {
				time.Sleep(time.Second * interval)
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

// Act invokes test according to appended Scenes.
func (x *Generalprobe) Act() error {
	for _, scene := range x.scenes {
		if err := scene.play(); err != nil {
			return err
		}
	}

	return nil
}
