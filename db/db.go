package db

import "embed"

//go:embed migrations/*.sql
var FSMigrations embed.FS
