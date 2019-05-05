package generalprobe

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

type Target interface {
	setGeneralprobe(gp *Generalprobe)
	arn() string
	name() string
}

type baseTarget struct {
	gp *Generalprobe
}

func (x *baseTarget) setGeneralprobe(gp *Generalprobe) { x.gp = gp }

// LogicalIDTarget is not expected to be controlled outside of generalprobe package.
// But it's exporeted just according to Go manner.
type LogicalIDTarget struct {
	baseTarget
	LogicalID string
}

// LogicalID is a target based on LogicalID of CloudFormation template.
func newLogicalID(name string) *LogicalIDTarget {
	return &LogicalIDTarget{LogicalID: name}
}

func (x *LogicalIDTarget) toArn(physicalID string) string {
	if len(strings.Split(physicalID, ":")) == 6 {
		return physicalID
	}

	type serviceHint struct {
		name   string
		prefix string
	}
	serviceMap := map[string]serviceHint{
		"AWS::Lambda::Function": serviceHint{"lambda", ""},
		"AWS::SNS::Topic":       serviceHint{"sns", ""},
		"AWS::DynamoDB::Table":  serviceHint{"dynamodb", "table/"},
		"AWS::Kinesis::Stream":  serviceHint{"kinesis", "stream/"},
	}

	resourceType := x.gp.LookupType(x.LogicalID)
	service, ok := serviceMap[resourceType]
	if !ok {
		log.WithFields(log.Fields{
			"logicalID":    x.LogicalID,
			"resourceType": resourceType,
		}).Fatal("The resource type is not supported")
	}

	return fmt.Sprintf("arn:aws:%s:%s:%s:%s%s", service.name, x.gp.awsRegion,
		x.gp.awsAccount, service.prefix, physicalID)
}

func (x *LogicalIDTarget) arn() string {
	pID := x.gp.LookupID(x.LogicalID)
	return x.toArn(pID)
}

func (x *LogicalIDTarget) name() string {
	pID := x.gp.LookupID(x.LogicalID)
	return pID
}

// ArnTarget is not expected to be controlled outside of generalprobe package.
// But it's exporeted just according to Go manner.
type ArnTarget struct {
	baseTarget
	arnData string
}

// Arn should be used to specify AWS resource out of CloudFormation template.
func newArn(arn string) *ArnTarget {
	sec := strings.Split(arn, ":")
	if len(sec) < 6 || 8 < len(sec) {
		log.WithField("arn", arn).Error("Invalid ARN format")
	}
	return &ArnTarget{arnData: arn}
}

func (x *ArnTarget) arn() string {
	return x.arnData
}

func (x *ArnTarget) name() string {
	// arn:partition:service:region:account-id:resource
	sec := strings.Split(x.arnData, ":")
	last := sec[len(sec)-1]
	resName := strings.Split(last, "/")
	if len(resName) == 2 {
		return resName[1]
	}

	return last
}
