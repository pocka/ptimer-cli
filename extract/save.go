// SPDX-FileCopyrightText: 2024 Shota FUJI <pockawoooh@gmail.com>
//
// SPDX-License-Identifier: Apache-2.0

package extract

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pocka/ptimer-cli/fileformat"

	"github.com/charmbracelet/log"
	_ "modernc.org/sqlite"
)

func retrieveMetadata(db *sql.DB) (*fileformat.Metadata, error) {
	rows, err := db.Query(`
		SELECT version, title, description, lang FROM metadata LIMIT 1;
	`)
	if err != nil {
		return nil, fmt.Errorf("Failed to read metadata record: %s", err)
	}

	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("No metadata record found")
	}

	var metadata fileformat.Metadata

	if err := rows.Scan(&metadata.Version, &metadata.Title, &metadata.Description, &metadata.Language); err != nil {
		return nil, err
	}

	return &metadata, nil
}

func retrieveSteps(db *sql.DB) ([]fileformat.Step, error) {
	rows, err := db.Query(`
		SELECT
			id, title, description,
			duration_seconds, "index", sound
		FROM step
		ORDER BY 'index' ASC;
	`)
	if err != nil {
		return nil, fmt.Errorf("Failed to read step records: %s", err)
	}

	defer rows.Close()

	var steps []fileformat.Step

	for rows.Next() {
		var step fileformat.Step

		if err := rows.Scan(
			&step.ID,
			&step.Title,
			&step.Description,
			&step.DurationSeconds,
			&step.Index,
			&step.Sound,
		); err != nil {
			return nil, fmt.Errorf("Failed to scan a step row: %s", err)
		}

		steps = append(steps, step)
	}

	return steps, nil
}

func retrieveAndSaveAssets(basePath string, outdir string, db *sql.DB) ([]fileformat.Asset, error) {
	rows, err := db.Query(`
		SELECT
			id, name, mime, data, notice
		FROM asset;
	`)
	if err != nil {
		return nil, fmt.Errorf("Failed to read asset records: %s", err)
	}

	defer rows.Close()

	var assets []fileformat.Asset

	for rows.Next() {
		var asset fileformat.Asset
		var data []byte

		if err := rows.Scan(
			&asset.ID,
			&asset.Name,
			&asset.MIME,
			&data,
			&asset.Notice,
		); err != nil {
			return nil, fmt.Errorf("Failed to scan an asset row: %s", err)
		}

		targetPath := filepath.Join(outdir, asset.Name)
		if matched, err := filepath.Match(outdir+string(filepath.Separator)+"*", targetPath); err != nil {
			return nil, fmt.Errorf("Unable to check asset write path boundary: %s", err)
		} else if !matched {
			return nil, fmt.Errorf(
				"Asset (id=%d) contains illegal parent traversal: tried to write to %s",
				asset.ID,
				targetPath,
			)
		}

		if err := os.WriteFile(targetPath, data, 0660); err != nil {
			return nil, fmt.Errorf("Failed to write asset (id=%d): %s", asset.ID, err)
		}

		relPath, err := filepath.Rel(basePath, targetPath)
		if err != nil {
			return nil, fmt.Errorf("Failed to generate relative path for an asset (id=%d): %s", asset.ID, err)
		}

		asset.Path = relPath

		assets = append(assets, asset)
	}

	return assets, nil
}

func readAndSave(dbPath string, outdir string) error {
	log.Debugf("Extract under %s", outdir)

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	jsonPath := filepath.Join(outdir, "ptimer.json")

	metadata, err := retrieveMetadata(db)
	if err != nil {
		return fmt.Errorf("Failed to retrieve metadata: %s", err)
	}

	steps, err := retrieveSteps(db)
	if err != nil {
		return fmt.Errorf("Failed to retrieve steps: %s", err)
	}

	assets, err := retrieveAndSaveAssets(filepath.Dir(jsonPath), outdir, db)
	if err != nil {
		return fmt.Errorf("Failed to retrieve assets: %s", err)
	}

	file := fileformat.Ptimer{
		Metadata: *metadata,
		Steps:    steps,
		Assets:   assets,
	}

	timerJson, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return fmt.Errorf("Failed to generate JSON data: %s", err)
	}

	if err := os.WriteFile(jsonPath, timerJson, 0660); err != nil {
		return fmt.Errorf("Failed to write JSON file: %s", err)
	}

	return nil
}
