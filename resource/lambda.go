package resource

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/pkg/errors"
)

// Lambda is resource entity for AWS Lambda.
type Lambda struct {
	// AWS SDK session to access lambda function
	Session *session.Session
	// FuncName is lambda function's explict resource name, not logical ID.
	FuncName string
}

// Invoke of Lambda is constructor of InvokeLambdaScene
func (x *Lambda) Invoke() *InvokeLambdaScene {
	return &InvokeLambdaScene{
		lambda: x,
	}
}

// InvokeLambdaScene is a test step to invoke Lambda Function and then invoke callback with response of Lambda.
type InvokeLambdaScene struct {
	lambda   *Lambda
	callback func(out []byte)
	payload  []byte
	input    lambda.InvokeInput
}

func (x *InvokeLambdaScene) Payload(data []byte) *InvokeLambdaScene {
	x.payload = data
	return x
}

func (x *InvokeLambdaScene) Callback(cb func(out []byte)) *InvokeLambdaScene {
	x.callback = cb
	return x
}

// SetInput copies original lambda.InvokeInput to own. FuncName and Payload should be overwritten if available.
func (x *InvokeLambdaScene) SetInput(input *lambda.InvokeInput) *InvokeLambdaScene {
	x.input = *input
	return x
}

// Name returns scene name
func (x *InvokeLambdaScene) Name() string {
	return fmt.Sprintf("Invoke Lambda: %s", x.lambda.FuncName)
}

// Play runs the test step
func (x *InvokeLambdaScene) Play() (bool, error) {
	client := lambda.New(x.lambda.Session)

	x.input.FunctionName = &x.lambda.FuncName
	if x.payload != nil {
		x.input.Payload = x.payload
	}

	output, err := client.Invoke(&x.input)
	if err != nil {
		return true, errors.Wrap(err, "Failed invoke Lambda")
	}

	if x.callback != nil {
		x.callback(output.Payload)
	}

	return true, nil
}
