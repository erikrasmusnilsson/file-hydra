package services

import "os"

// FileExists returns true if a file exists
// under the given path fn.
func FileExists(fn string) bool {
	_, err := os.Stat(fn)
	return err == nil
}
