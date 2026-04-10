//go:build go1.25

package synctestx

import (
	"testing/synctest"
)

var Run = synctest.Test
