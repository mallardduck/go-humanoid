package humanoid_go

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSpaceIdGenerator(t *testing.T) {
	spaceHumanoid, err := SpaceIdGenerator()
	assert.NoError(t, err)
	res, err := spaceHumanoid.Create(0)
	assert.NoError(t, err)
	assert.Exactly(t, "andromeda", res)
}
