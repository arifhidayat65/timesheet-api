package db

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Embed semua file .sql di internal/db/migrations/ (path relatif ke file ini)
//go:embed migrations/*.sql
var embeddedMigrations embed.FS

// Migrate akan:
// 1) Jika ENV MIGRATIONS_DIR diset → baca dari folder itu (untuk dev/override).
// 2) Jika tidak → pakai file yang di-embed.
func Migrate(db *sql.DB) error {
	if dir := strings.TrimSpace(os.Getenv("MIGRATIONS_DIR")); dir != "" {
		return runFromDir(db, dir)
	}
	return runFromFS(db, embeddedMigrations, "migrations")
}

func runFromFS(db *sql.DB, efs fs.FS, root string) error {
	ents, err := fs.ReadDir(efs, root)
	if err != nil {
		return err
	}
	var names []string
	for _, e := range ents {
		if e.IsDir() { continue }
		name := e.Name()
		if strings.HasSuffix(strings.ToLower(name), ".sql") {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	for _, name := range names {
		b, err := fs.ReadFile(efs, filepath.Join(root, name))
		if err != nil { return err }
		if _, err := db.Exec(string(b)); err != nil {
			return fmt.Errorf("migrate %s: %w", name, err)
		}
	}
	return nil
}

func runFromDir(db *sql.DB, dir string) error {
	ents, err := os.ReadDir(dir)
	if err != nil { return err }
	var paths []string
	for _, e := range ents {
		if e.IsDir() { continue }
		if strings.HasSuffix(strings.ToLower(e.Name()), ".sql") {
			paths = append(paths, filepath.Join(dir, e.Name()))
		}
	}
	sort.Strings(paths)
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil { return err }
		if _, err := db.Exec(string(b)); err != nil {
			return fmt.Errorf("migrate %s: %w", p, err)
		}
	}
	return nil
}
