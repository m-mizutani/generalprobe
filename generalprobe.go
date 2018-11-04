package generalprobe

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

type Generalprobe struct {
	awsRegion string
	stackName string
	scenes    []Scene
	resources []*cloudformation.StackResource
	done      bool
}

func New(awsRegion, stackName string) Generalprobe {
	gp := Generalprobe{
		awsRegion: awsRegion,
		stackName: stackName,
		done:      false,
	}

	ssn := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(awsRegion),
	}))
	client := cloudformation.New(ssn)

	resp, err := client.DescribeStackResources(&cloudformation.DescribeStackResourcesInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		log.Fatal("Fail to get CloudFormation Stack resources: ", err)
	}

	gp.resources = resp.StackResources

	return gp
}

func (x *Generalprobe) lookup(logicalId string) string {
	for _, resource := range x.resources {
		if resource.LogicalResourceId != nil && *resource.LogicalResourceId == logicalId {
			return *resource.PhysicalResourceId
		}
	}

	return ""
}

func (x *Generalprobe) AddScenes(newScenes []Scene) {
	for _, scene := range newScenes {
		scene.setGeneralprobe(x)
		x.scenes = append(x.scenes, scene)
	}
}

func (x *Generalprobe) Act() error {
	for _, scene := range x.scenes {
		if err := scene.play(); err != nil {
			return err
		}
	}

	return nil
}
