// SPDX-FileCopyrightText: 2024 Shota FUJI <pockawoooh@gmail.com>
//
// SPDX-License-Identifier: Apache-2.0

package fileformat

type Metadata struct {
	Version     string  `json:"version"`
	Title       string  `json:"title"`
	Description *string `json:"description"`
	Language    string  `json:"lang"`
}

type Step struct {
	ID              uint    `json:"id"`
	Title           string  `json:"title"`
	Description     *string `json:"description"`
	Sound           *uint   `json:"sound"`
	DurationSeconds *uint   `json:"duration_seconds"`
	Index           int     `json:"index"`
}

type Asset struct {
	ID     uint    `json:"id"`
	Name   string  `json:"name"`
	MIME   string  `json:"mime"`
	Path   string  `json:"path"`
	Notice *string `json:"notice"`
}

type Ptimer struct {
	Metadata Metadata `json:"metadata"`
	Steps    []Step   `json:"steps"`
	Assets   []Asset  `json:"assets"`
}
