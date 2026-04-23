package synctestx_test

import (
	"testing"

	"go.atoms.co/lib/testing/synctestx"
)

func TestRun(t *testing.T) {
	synctestx.Run(t, "run", func(t *testing.T) {
	})
}
