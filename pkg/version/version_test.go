package version

import (
	"fmt"
	"testing"
)

func TestFullVersionName(t *testing.T) {
	origVersion := ProviderVersion
	origTag := TagSuffix
	t.Cleanup(func() {
		ProviderVersion = origVersion
		TagSuffix = origTag
	})

	ProviderVersion = "1.2.3"
	TagSuffix = "stable"

	expected := fmt.Sprintf("v%s-%s", ProviderVersion, TagSuffix)
	got := FullVersionName()
	if got != expected {
		t.Errorf("FullVersionName() = %q, want %q", got, expected)
	}
}

func TestFullVersionName_DefaultValues(t *testing.T) {
	got := FullVersionName()
	expected := fmt.Sprintf("v%s-%s", ProviderVersion, TagSuffix)
	if got != expected {
		t.Errorf("FullVersionName() = %q, want %q", got, expected)
	}
}
