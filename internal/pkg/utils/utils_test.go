package utils

import (
	"strings"
	"testing"
)

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

func TestAddr(t *testing.T) {
	valoper := "terravaloper1khfcg09plqw84jxy5e7fj6ag4s2r9wqsgm7k94"
	want := "terra1khfcg09plqw84jxy5e7fj6ag4s2r9wqsg5jt4x"
	got, err := ValoperToAccAddress(valoper)
	if err != nil {
		t.Fatalf("error is not nil: %v\n", err)
	}

	if got != want {
		t.Fatalf("got \"%s\" but want \"%s\"\n", got, want)
	}

	incorrectValoper := "terravaloper1khfcg09plqw84jxy5e7fj6ag4s2r9wqsgm7k95"
	_, err = ValoperToAccAddress(incorrectValoper)
	if err == nil {
		t.Fatalf("error is nil\n")
	}
}
