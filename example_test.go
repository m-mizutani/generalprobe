package generalprobe_test

import (
	"testing"

	"github.com/m-mizutani/generalprobe"
)

func TestExamplePlayBook(t *testing.T) {
	stack := generalprobe.Stack{
		StackName: "SampleStack",
		Region:    "ap-northeast-1",
	}

	scenes := []generalprobe.Scene{
		stack.SNS("test").PublishMessage("xxx"),
	}

	playbook := &generalprobe.PlayBook{
		Scenes: scenes,
	}

	err := playbook.Play()
	if err != nil {
		t.Errorf("Failed playbook: %w", err)
	}
}
