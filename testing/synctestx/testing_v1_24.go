//go:build !go1.25

package synctestx

import (
	"testing"
	"testing/synctest"
)

// Run executes a test in synctest context
func Run(t *testing.T, name string, f func(t *testing.T)) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		synctest.Run(func() {
			f(t)
		})
	})
}
