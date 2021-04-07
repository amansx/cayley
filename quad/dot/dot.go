// Package dot is deprecated. Use github.com/amansx/quad/dot.
package dot

import (
	"io"

	"github.com/amansx/quad/dot"
)

func NewWriter(w io.Writer) *Writer {
	return dot.NewWriter(w)
}

type Writer = dot.Writer
