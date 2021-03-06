package generalprobe

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"

	// "github.com/k0kubun/pp"
	"github.com/pkg/errors"
)

// InvokeLambdaScene is a scene only to invoke AWS Lambda.
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
func InvokeLambda(target Target, callback InvokeLambdaCallback) *InvokeLambdaScene {
	scene := InvokeLambdaScene{
		target:   target,
		callback: callback,
	}
	return &scene
}

// Strings return text explanation of the scene
func (x *InvokeLambdaScene) string() string {
	return fmt.Sprintf("Invoke Lambda %s", x.target.arn(x.gp))
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

// SnsEvent sets SNS event as argument of invoke Lambda.
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

// Event sets general event structure as argument of invoke Lambda.
// event will be marshaled to JSON string and pass it to Lambda.
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

	lambdaArn := x.target.arn(x.gp)
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
