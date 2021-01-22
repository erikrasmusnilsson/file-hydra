package services

import (
	"os"
	"testing"
)

func TestFileExists(t *testing.T) {
	n := "file.test"
	os.Create(n)
	if !FileExists(n) {
		t.Errorf("Expected FileExists(%s) to return true, returned false.", n)
		return
	}
	os.Remove(n)
	if FileExists(n) {
		t.Errorf("Expected FileExists(%s) to return false, returned true.", n)
	}
}
