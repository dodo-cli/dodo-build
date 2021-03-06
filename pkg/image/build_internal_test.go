package image

import (
	"testing"

	"github.com/dodo-cli/dodo-build/pkg/types"
	"github.com/moby/buildkit/client"
	"github.com/stretchr/testify/assert"
)

func TestBuildImage(t *testing.T) {
	displayCh := make(chan *client.SolveStatus)
	defer close(displayCh)

	image := fakeImage(t, &types.BuildInfo{
		Context: "./test",
	})
	result, err := image.runBuild(&contextData{
		remote:         "client-session",
		dockerfileName: "Dockerfile",
	}, displayCh)
	assert.Nil(t, err)
	assert.Equal(t, "NewImageID", result)
}

func TestBuildInlineImage(t *testing.T) {
	displayCh := make(chan *client.SolveStatus)
	defer close(displayCh)

	image := fakeImage(t, &types.BuildInfo{
		InlineDockerfile: []string{"FROM scratch"},
	})
	result, err := image.runBuild(&contextData{
		remote:         "client-session",
		dockerfileName: "Dockerfile",
	}, displayCh)
	assert.Nil(t, err)
	assert.Equal(t, "NewImageID", result)
}
