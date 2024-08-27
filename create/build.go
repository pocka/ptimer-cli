// SPDX-FileCopyrightText: 2024 Shota FUJI <pockawoooh@gmail.com>
//
// SPDX-License-Identifier: Apache-2.0

package create

import (
	_ "embed"

	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pocka/ptimer-cli/fileformat"

	_ "modernc.org/sqlite"
)

// go:embed init.sql
var initSQL string

func buildPtimerFile(data fileformat.Ptimer, basePath string, dbPath string) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("Failed to open database: %s", err)
	}
	defer db.Close()

	if _, err := db.Exec(initSQL); err != nil {
		return fmt.Errorf("Failed to initialize database: %s", err)
	}

	_, err = db.Exec(`
		INSERT OR ABORT INTO metadata (
			version,
			title,
			description,
			lang
		) VALUES (
			?, ?, ?, ?
		);
	`, data.Metadata.Version, data.Metadata.Title, data.Metadata.Description, data.Metadata.Language)
	if err != nil {
		return fmt.Errorf("Failed to write metadata: %s", err)
	}

	stepInsertStmt, err := db.Prepare(`
		INSERT OR ABORT INTO step (
			id,
			title,
			description,
			sound,
			duration_seconds,
			'index'
		) VALUES (
			?, ?, ?, ?, ?, ?
		);
	`)
	if err != nil {
		return fmt.Errorf("Failed to prepare step creation statement: %s", err)
	}
	defer stepInsertStmt.Close()

	for _, step := range data.Steps {
		if _, err := stepInsertStmt.Exec(
			step.ID,
			step.Title,
			step.Description,
			step.Sound,
			step.DurationSeconds,
			step.Index,
		); err != nil {
			return fmt.Errorf("Failed to write step (id=%d): %s", step.ID, err)
		}
	}

	assetInsertStmt, err := db.Prepare(`
		INSERT OR ABORT INTO asset (
			id,
			name,
			mime,
			data,
			notice
		) VALUES (
			?, ?, ?, ?, ?
		);
	`)
	if err != nil {
		return fmt.Errorf("Failed to prepare asset creation statement: %s", err)
	}
	defer assetInsertStmt.Close()

	for _, asset := range data.Assets {
		binaryPath := filepath.Join(basePath, asset.Path)
		if matched, err := filepath.Match(basePath+string(filepath.Separator)+"*", binaryPath); err != nil {
			return fmt.Errorf("Failed to read asset file: %s", err)
		} else if !matched {
			return fmt.Errorf("Failed to read asset file: asset.path cannot refer outside the directory JSON file is in")
		}

		binary, err := os.ReadFile(binaryPath)
		if err != nil {
			return fmt.Errorf("Failed to read asset file: %s", err)
		}

		if _, err := assetInsertStmt.Exec(
			asset.ID,
			asset.Name,
			asset.MIME,
			binary,
			asset.Notice,
		); err != nil {
			return fmt.Errorf("Failed to write asset (id=%d): %s", asset.ID, err)
		}
	}

	return nil
}
