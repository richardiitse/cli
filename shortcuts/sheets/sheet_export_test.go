// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package sheets

import (
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/larksuite/cli/internal/cmdutil"
	"github.com/larksuite/cli/internal/httpmock"
)

func TestSheetExportRejectsOverwriteWithoutFlag(t *testing.T) {
	f, _, _, _ := cmdutil.TestFactory(t, sheetsTestConfig())

	tmpDir := t.TempDir()
	cmdutil.TestChdir(t, tmpDir)

	if err := os.WriteFile("report.xlsx", []byte("old"), 0644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	// The overwrite check happens before any API call, so no HTTP stubs needed.
	err := mountAndRunSheets(t, SheetExport, []string{
		"+export",
		"--spreadsheet-token", "shtTOKEN",
		"--file-extension", "xlsx",
		"--output-path", "report.xlsx",
		"--as", "user",
	}, f, nil)
	if err == nil {
		t.Fatal("expected overwrite protection error, got nil")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSheetExportAllowsOverwriteWithFlag(t *testing.T) {
	f, _, _, reg := cmdutil.TestFactory(t, sheetsTestConfig())

	// Register stubs for the export task creation API.
	reg.Register(&httpmock.Stub{
		Method: "POST",
		URL:    "/open-apis/drive/v1/export_tasks",
		Body: map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{"ticket": "tkt_123"},
		},
	})
	// Register stub for the poll API (returns completed immediately).
	reg.Register(&httpmock.Stub{
		Method: "GET",
		URL:    "/open-apis/drive/v1/export_tasks/tkt_123",
		Body: map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"result": map[string]interface{}{
					"file_token": "box_export_123",
					"file_name":  "report.xlsx",
					"file_size":  100,
				},
			},
		},
	})
	// Register stub for the download API.
	reg.Register(&httpmock.Stub{
		Method:  "GET",
		URL:     "/open-apis/drive/v1/export_tasks/file/box_export_123/download",
		RawBody: []byte("new-content"),
		Headers: http.Header{
			"Content-Type": []string{"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
		},
	})

	tmpDir := t.TempDir()
	cmdutil.TestChdir(t, tmpDir)

	if err := os.WriteFile("report.xlsx", []byte("old"), 0644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	err := mountAndRunSheets(t, SheetExport, []string{
		"+export",
		"--spreadsheet-token", "shtTOKEN",
		"--file-extension", "xlsx",
		"--output-path", "report.xlsx",
		"--overwrite",
		"--as", "user",
	}, f, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile("report.xlsx")
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}
	if string(data) != "new-content" {
		t.Fatalf("file content = %q, want %q", string(data), "new-content")
	}
}

func TestSheetExportNoOutputPathSkipsOverwriteCheck(t *testing.T) {
	// When --output-path is not provided, the overwrite check should be
	// skipped entirely (no file-exists error even if cwd has a file with
	// the same name). We only need to verify the early validation phase
	// passes — the rest of Execute calls export APIs which are unrelated
	// to the overwrite feature.
	f, _, _, reg := cmdutil.TestFactory(t, sheetsTestConfig())

	reg.Register(&httpmock.Stub{
		Method: "POST",
		URL:    "/open-apis/drive/v1/export_tasks",
		Body: map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{"ticket": "tkt_456"},
		},
	})
	reg.Register(&httpmock.Stub{
		Method: "GET",
		URL:    "/open-apis/drive/v1/export_tasks/tkt_456",
		Body: map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"result": map[string]interface{}{
					"file_token": "box_export_456",
				},
			},
		},
	})
	// The code falls through to download even without --output-path,
	// so we need a download stub to avoid an unmatched-stub verify failure.
	reg.Register(&httpmock.Stub{
		Method:  "GET",
		URL:     "/open-apis/drive/v1/export_tasks/file/box_export_456/download",
		RawBody: []byte("data"),
	})

	tmpDir := t.TempDir()
	cmdutil.TestChdir(t, tmpDir)

	// Create a file that would collide if overwrite check were applied.
	if err := os.WriteFile("report.xlsx", []byte("old"), 0644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	err := mountAndRunSheets(t, SheetExport, []string{
		"+export",
		"--spreadsheet-token", "shtTOKEN",
		"--file-extension", "xlsx",
		"--as", "user",
	}, f, nil)
	// The error here is from the download phase (SafeOutputPath("") fails),
	// not from overwrite protection. Verify no "already exists" error.
	if err != nil && strings.Contains(err.Error(), "already exists") {
		t.Fatalf("overwrite check should be skipped when --output-path is empty, got: %v", err)
	}
}
