//+build cgo

package all

import (
	// backends requiring cgo
	_ "github.com/amansx/cayley/graph/sql/sqlite"
)
