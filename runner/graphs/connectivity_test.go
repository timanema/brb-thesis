package graphs

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFindConnectedness1Gen(t *testing.T) {
	n, k := 10, 3
	m := GeneralizedWheelGenerator{}

	g, err := m.Generate(n, k, 0)
	assert.NoError(t, err)
	assert.Equal(t, k, FindConnectedness(g))
}

func TestFindConnectedness2Gen(t *testing.T) {
	n, k := 50, 30
	m := GeneralizedWheelGenerator{}

	g, err := m.Generate(n, k, 0)
	assert.NoError(t, err)
	assert.Equal(t, k, FindConnectedness(g))
}

func TestFindConnectedness3Gen(t *testing.T) {
	n, k := 150, 40
	m := GeneralizedWheelGenerator{}

	g, err := m.Generate(n, k, 0)
	assert.NoError(t, err)
	assert.Equal(t, k, FindConnectedness(g))
}

func TestFindConnectedness1Mul(t *testing.T) {
	n, k := 5, 2
	m := MultiPartiteWheelGenerator{}

	g, err := m.Generate(n, k, 0)
	assert.NoError(t, err)
	assert.Equal(t, k, FindConnectedness(g))
}

func TestFindConnectedness2Mul(t *testing.T) {
	n, k := 25, 10
	m := MultiPartiteWheelGenerator{}

	g, err := m.Generate(n, k, 0)
	assert.NoError(t, err)
	assert.Equal(t, k, FindConnectedness(g))
}

func TestFindConnectedness3Mul(t *testing.T) {
	n, k := 50, 24
	m := MultiPartiteWheelGenerator{}

	g, err := m.Generate(n, k, 0)
	assert.NoError(t, err)
	assert.Equal(t, k, FindConnectedness(g))
}
