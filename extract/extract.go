// SPDX-FileCopyrightText: 2024 Shota FUJI <pockawoooh@gmail.com>
//
// SPDX-License-Identifier: Apache-2.0

package extract

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	var outdir string
	var clean bool

	cmd := &cobra.Command{
		Use:           "extract [ptimer file]",
		Short:         "Save .ptimer contents into standard files",
		Long:          "Extracts timer data and asset files then save it under a specified directory.",
		Args:          cobra.MaximumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if outdir == "" {
				return fmt.Errorf("--outdir flag is required.")
			}

			outdir, err := filepath.Abs(outdir)
			if err != nil {
				return fmt.Errorf("Failed to resolve outdir: %s", err)
			}

			if _, err := os.Stat(outdir); err != nil {
				if !os.IsNotExist(err) {
					return fmt.Errorf("Unable to open the outdir: %s", err)
				}
			} else {
				if !clean {
					return fmt.Errorf("Cannot create output directory at %s: file or directory exists", outdir)
				}

				if err := os.RemoveAll(outdir); err != nil {
					return err
				}
			}

			if err := os.Mkdir(outdir, 0700); err != nil {
				return err
			}

			var ptimerFilepath string
			if len(args) == 1 {
				ptimerFilepath, err = filepath.Abs(args[0])

				if err != nil {
					return fmt.Errorf("Path for .ptimer file is invalid: %s", err)
				}
			} else {
				file, err := os.CreateTemp("", "stdin.ptimer")
				if err != nil {
					return err
				}
				defer file.Close()
				defer os.Remove(file.Name())

				if _, err := io.Copy(file, os.Stdin); err != nil {
					return err
				}

				ptimerFilepath = file.Name()
			}

			if err := readAndSave(ptimerFilepath, outdir); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&outdir, "outdir", "", "Path to the output directory.")
	cmd.Flags().BoolVar(
		&clean,
		"clean",
		false,
		`Whether to remove an output directory before extracting.
If this flag is not set and the output directory is not empty, this program exit with error code.`,
	)

	return cmd
}
