package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValConsPub(t *testing.T) {
	valcons := "terravalcons1qw4gg8v3jt0tfaq2qv337sa22slj3dxu73tyql"
	valconspub := "terravalconspub1zcjduepq5zcrunelz9yy09ksug5tcvx7r46mslnxk9gxqp8xflmwm3md8aesw6u3a8"

	expectedValConsAddr, err := ValConsToAddr(valcons)
	assert.NoError(t, err)
	actualValConsAddr, err := ValConsPubToAddr(valconspub)
	assert.NoError(t, err)
	assert.Equal(t, expectedValConsAddr, actualValConsAddr)
}

func TestTerraMonitorsPath(t *testing.T) {
	githubDir := "/home/runner/work/terra-monitors/terra-monitors/internal/app/collector/monitors"
	expected := "/home/runner/work/terra-monitors/terra-monitors/tests/"

	path := getTerraMonitorsPath(githubDir)

	if !strings.Contains(path, expected) {
		t.Errorf("expected: %s, got: %s", expected, path)
	}

	if path != expected {
		t.Errorf("expected: %s, got: %s", expected, path)
	}

	localDir := "/Users/callmepak/go/src/terra-monitors/internal/app/collector/monitors"
	expected = "/Users/callmepak/go/src/terra-monitors/tests/"

	path = getTerraMonitorsPath(localDir)

	if !strings.Contains(path, expected) {
		t.Errorf("expected: %s, got: %s", expected, path)
	}
}
