package generalprobe

import (
	"reflect"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"

	"github.com/sirupsen/logrus"
)

func init() {
	setupLogger()
}

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
func New(awsRegion, stackName string) *Generalprobe {
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

	return &gp
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

// Play executes defined scenes sequentially.
func (x *Generalprobe) Play(playbook []Scene) error {
	for idx, scene := range playbook {
		scene.setGeneralprobe(x)
		logger.Infof("Step (%d/%d): %s (%s)\n", idx+1, len(playbook), scene.string(), reflect.TypeOf(scene))

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
