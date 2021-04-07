// Package graphml is deprecated. Use github.com/amansx/quad/graphml.
package graphml

import (
	"io"

	"github.com/amansx/quad/graphml"
)

func NewWriter(w io.Writer) *Writer {
	return graphml.NewWriter(w)
}

type Writer = graphml.Writer
