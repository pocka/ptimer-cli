// SPDX-FileCopyrightText: 2024 Shota FUJI <pockawoooh@gmail.com>
//
// SPDX-License-Identifier: Apache-2.0

package create

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pocka/ptimer-cli/fileformat"

	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	var outPath string

	cmd := &cobra.Command{
		Use:           "create",
		Short:         "Generate .ptimer file from JSON and asset files",
		Args:          cobra.MaximumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			var jsonData []byte
			var basePath string
			if len(args) == 1 {
				jsonFilepath, err := filepath.Abs(args[0])
				if err != nil {
					return fmt.Errorf("Path for a JSON file is invalid: %s", err)
				}

				jsonData, err = os.ReadFile(jsonFilepath)
				basePath = filepath.Dir(jsonFilepath)
			} else {
				str, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("Failed to read from stdin: %s", err)
				}

				jsonData = str
				basePath, err = os.Getwd()
				if err != nil {
					return fmt.Errorf("Failed to get working directory: %s", err)
				}
			}

			var data fileformat.Ptimer

			if err := json.Unmarshal(jsonData, &data); err != nil {
				return fmt.Errorf("Failed to parse input JSON file: %s", err)
			}

			if outPath == "" {
				file, err := os.CreateTemp("", "stdout.ptimer")
				if err != nil {
					return fmt.Errorf("Failed to create temporary file: %s", err)
				}
				defer os.Remove(file.Name())
				if err := file.Close(); err != nil {
					return fmt.Errorf("Failed to close temporary file: %s", err)
				}

				if err := buildPtimerFile(data, basePath, file.Name()); err != nil {
					return fmt.Errorf("Failed to build .ptimer file: %s", err)
				}

				file, err = os.Open(file.Name())
				if err != nil {
					return fmt.Errorf("Failed to open temporary file: %s", err)
				}
				defer file.Close()

				_, err = io.Copy(os.Stdout, file)

				return err
			}

			if _, err := os.Stat(outPath); err == nil {
				if err := os.Remove(outPath); err != nil {
					return fmt.Errorf("Failed to clean target file: %s", err)
				}
			}

			return buildPtimerFile(data, basePath, outPath)
		},
	}

	cmd.Flags().StringVar(
		&outPath,
		"out",
		"",
		`Path to the output file.
If this flag is not set, create command the resulting file to stdout.`,
	)

	return cmd
}
