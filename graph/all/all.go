package all

import (
	// supported backends
	_ "github.com/amansx/cayley/graph/kv/all"
	_ "github.com/amansx/cayley/graph/memstore"
	_ "github.com/amansx/cayley/graph/nosql/all"
	_ "github.com/amansx/cayley/graph/sql/cockroach"
	_ "github.com/amansx/cayley/graph/sql/mysql"
	_ "github.com/amansx/cayley/graph/sql/postgres"
)
