package migrations

import "embed"

// FS contains the ordered SQL migration files.
//
//go:embed *.sql
var FS embed.FS
