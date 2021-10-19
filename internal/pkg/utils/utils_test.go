package utils

import (
	"strings"
	"testing"
)

func TestTerraMonitorsPath(t *testing.T) {
	githubDir := "/home/runner/work/terra-monitors/terra-monitors/internal/app/collector/monitors"
	expected := "/home/runner/work/terra-monitors/terra-monitors/tests/"

	path, err := getTerraMonitorsPath(githubDir)
	if err != nil {
		t.Error(err)
	}

	if !strings.Contains(path, expected) {
		t.Errorf("expected: %s, got: %s", expected, path)
	}

	if path != expected {
		t.Errorf("expected: %s, got: %s", expected, path)
	}

	localDir := "/Users/callmepak/go/src/terra-monitors/internal/app/collector/monitors"
	expected = "/Users/callmepak/go/src/terra-monitors/tests/"

	path, err = getTerraMonitorsPath(localDir)
	if err != nil {
		t.Error(err)
	}

	if !strings.Contains(path, expected) {
		t.Errorf("expected: %s, got: %s", expected, path)
	}
}
