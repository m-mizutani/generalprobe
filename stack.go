package generalprobe

import (
	"io"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/m-mizutani/generalprobe/resource"
	"github.com/pkg/errors"
)

type Stack struct {
	StackName string
	Region    string

	manifest       *cdkoutManifest
	stackResources []*cloudformation.StackResource
}

func (x *Stack) session() *session.Session {
	return session.New(aws.NewConfig().WithRegion(x.Region))
}

func (x *Stack) LoadCdkManifest(rd io.Reader) error {
	manifest, err := newManifest(rd)
	if err != nil {
		return errors.Wrap(err, "Failed to newManifest")
	}

	x.manifest = manifest
	return nil
}

func (x *Stack) lookupStackResource(id string) (*cloudformation.StackResource, error) {
	logicalID := id
	if x.manifest != nil {
		resourceID, err := x.manifest.lookupLogicalID(x.StackName, id)
		if err != nil {
			return nil, err
		}
		logicalID = *resourceID
	}

	if err := x.loadStackResources(); err != nil {
		return nil, errors.Wrap(err, "Failed to loadStackResource")
	}

	for i := 0; i < len(x.stackResources); i++ {
		if aws.StringValue(x.stackResources[i].LogicalResourceId) == logicalID {
			return x.stackResources[i], nil
		}
	}

	return nil, nil
}

func (x *Stack) loadStackResources() error {
	if x.stackResources != nil { // Already loaded
		return nil
	}

	cfn := cloudformation.New(session.New(), aws.NewConfig().WithRegion(x.Region))

	req := cloudformation.DescribeStackResourcesInput{
		StackName: aws.String(x.StackName),
	}
	resp, err := cfn.DescribeStackResources(&req)
	if err != nil {
		return err
	}

	x.stackResources = resp.StackResources
	return nil
}

// Resource generators

// SNS creates a new SNS entity by stack resource
func (x *Stack) SNS(logicalID string) *resource.SNS {
	r, err := x.lookupStackResource(logicalID)
	if err != nil {
		log.Fatalf("Error in looking up %s in %s: %v", logicalID, x.StackName, err)
	}
	if r == nil {
		log.Fatalf("%s is not found in %s", logicalID, x.StackName)
	}

	return &resource.SNS{
		TopicARN: aws.StringValue(r.PhysicalResourceId),
		Session:  x.session(),
	}
}

// Lambda creates a new Lambda entity by stack resource
func (x *Stack) Lambda(logicalID string) *resource.Lambda {
	r, err := x.lookupStackResource(logicalID)
	if err != nil {
		log.Fatalf("Error in looking up %s in %s: %v", logicalID, x.StackName, err)
	}
	if r == nil {
		log.Fatalf("%s is not found in %s", logicalID, x.StackName)
	}

	return &resource.Lambda{
		FuncName: aws.StringValue(r.PhysicalResourceId),
		Session:  x.session(),
	}
}

// SQS creates a new SQS entity by stack resource
func (x *Stack) SQS(logicalID string) *resource.SQS {
	r, err := x.lookupStackResource(logicalID)
	if err != nil {
		log.Fatalf("Error in looking up %s in %s: %v", logicalID, x.StackName, err)
	}
	if r == nil {
		log.Fatalf("%s is not found in %s", logicalID, x.StackName)
	}

	return &resource.SQS{
		QueueURL: aws.StringValue(r.PhysicalResourceId),
		Session:  x.session(),
	}
}

// DynamoDB creates a new DynamoDB entity by stack resource
func (x *Stack) DynamoDB(logicalID string) *resource.DynamoDB {
	r, err := x.lookupStackResource(logicalID)
	if err != nil {
		log.Fatalf("Error in looking up %s in %s: %v", logicalID, x.StackName, err)
	}
	if r == nil {
		log.Fatalf("%s is not found in %s", logicalID, x.StackName)
	}

	return &resource.DynamoDB{
		TableName: aws.StringValue(r.PhysicalResourceId),
		Session:   x.session(),
	}
}
