package humanoid_go

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSpaceIdGenerator(t *testing.T) {
	categories := make([]string, 4)
	categories[0] = "colors"
	categories[1] = "buzzwords"
	categories[2] = "life-cycle"
	categories[3] = "star-taxonomy"
	spaceHumanoid, err := SpaceIdGenerator()
	assert.NoError(t, err)

	res, err := spaceHumanoid.Create(0)
	assert.NoError(t, err)
	assert.Exactly(t, "andromeda", res)

	res, err = spaceHumanoid.Create(1)
	assert.NoError(t, err)
	assert.Exactly(t, "backward", res)

	res, err = spaceHumanoid.Create(2)
	assert.NoError(t, err)
	assert.Exactly(t, "bode", res)

	res, err = spaceHumanoid.Create(3)
	assert.NoError(t, err)
	assert.Exactly(t, "cigar", res)

	res, err = spaceHumanoid.Create(23)
	assert.NoError(t, err)
	assert.Exactly(t, "eris-pinwheel", res)
}
