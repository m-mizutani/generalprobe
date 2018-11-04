package generalprobe_test

import (
	"testing"

	gp "github.com/m-mizutani/generalprobe"
	"github.com/stretchr/testify/assert"
)

func TestBasicUsage(t *testing.T) {
	g := gp.New("ap-northeast-1", "chamber-test")
	n := 0
	g.AddScenes([]gp.Scene{
		gp.Intermission(func() {
			n++
		}),
	})
	g.Act()
	assert.Equal(t, 1, n)
}
