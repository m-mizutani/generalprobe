package generalprobe

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"

	// "github.com/k0kubun/pp"
	"github.com/pkg/errors"
)

type InvokeLambdaScene struct {
	target   Target
	input    []byte
	event    interface{}
	callback InvokeLambdaCallback
	baseScene
}

// InvokeLambdaCallback is callback function called after Lambda exits
type InvokeLambdaCallback func(response []byte)

// InvokeLambda is a constructor of Scene
func (x *Generalprobe) InvokeLambda(target Target, callback InvokeLambdaCallback) *InvokeLambdaScene {
	scene := InvokeLambdaScene{
		target:   target,
		callback: callback,
	}
	return &scene
}

func toMessage(msg interface{}) string {
	switch v := msg.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		raw, err := json.Marshal(v)
		if err != nil {
			logger.WithField("message", msg).Fatal(err)
		}
		return string(raw)
	}
}

func (x *InvokeLambdaScene) SnsEvent(input interface{}) *InvokeLambdaScene {
	msg := toMessage(input)
	event := events.SNSEvent{
		Records: []events.SNSEventRecord{
			events.SNSEventRecord{
				SNS: events.SNSEntity{
					Message: msg,
				},
			},
		},
	}
	return x.Event(event)
}

func (x *InvokeLambdaScene) Event(event interface{}) *InvokeLambdaScene {
	x.event = event
	return x
}

func (x *InvokeLambdaScene) play() error {
	eventData, err := json.Marshal(x.event)
	if err != nil {
		return errors.Wrap(err, "unmarshal event")
	}

	ssn := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(x.region()),
	}))
	lambdaService := lambda.New(ssn)

	lambdaArn := x.target.arn()
	resp, err := lambdaService.Invoke(&lambda.InvokeInput{
		FunctionName: aws.String(lambdaArn),
		Payload:      eventData,
	})
	if err != nil {
		logger.Fatal("Fail to invoke lambda", err)
	}

	logger.WithField("response", resp).Debug("lamba invoked")

	x.callback(resp.Payload)

	return nil
}
