// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package sheets

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/larksuite/cli/internal/cmdutil"
	"github.com/larksuite/cli/internal/core"
	"github.com/larksuite/cli/internal/httpmock"
	"github.com/larksuite/cli/shortcuts/common"
)

func sheetsTestConfig() *core.CliConfig {
	return &core.CliConfig{
		AppID: "sheets-test-app", AppSecret: "test-secret", Brand: core.BrandFeishu,
	}
}

func mountAndRunSheets(t *testing.T, s common.Shortcut, args []string, f *cmdutil.Factory, stdout *bytes.Buffer) error {
	t.Helper()
	parent := &cobra.Command{Use: "sheets"}
	s.Mount(parent, f)
	parent.SetArgs(args)
	parent.SilenceErrors = true
	parent.SilenceUsage = true
	if stdout != nil {
		stdout.Reset()
	}
	return parent.Execute()
}

// ── Validate ─────────────────────────────────────────────────────────────────

func TestSheetWriteImageValidateRequiresToken(t *testing.T) {
	t.Parallel()
	runtime := newSheetsTestRuntime(t, map[string]string{
		"image": "./logo.png",
		"range": "A1",
	}, nil)
	err := SheetWriteImage.Validate(context.Background(), runtime)
	if err == nil || !strings.Contains(err.Error(), "--url or --spreadsheet-token") {
		t.Fatalf("expected token error, got: %v", err)
	}
}

func TestSheetWriteImageValidateAcceptsURL(t *testing.T) {
	t.Parallel()
	runtime := newSheetsTestRuntime(t, map[string]string{
		"url":      "https://example.larksuite.com/sheets/shtABC123",
		"image":    "./logo.png",
		"range":    "sheetId!A1:A1",
		"sheet-id": "",
	}, nil)
	err := SheetWriteImage.Validate(context.Background(), runtime)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSheetWriteImageValidateAcceptsSpreadsheetToken(t *testing.T) {
	t.Parallel()
	runtime := newSheetsTestRuntime(t, map[string]string{
		"spreadsheet-token": "shtABC123",
		"image":             "./logo.png",
		"range":             "sheetId!A1:A1",
		"sheet-id":          "",
	}, nil)
	err := SheetWriteImage.Validate(context.Background(), runtime)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ── DryRun ───────────────────────────────────────────────────────────────────

func TestSheetWriteImageDryRun(t *testing.T) {
	t.Parallel()
	runtime := newSheetsTestRuntime(t, map[string]string{
		"spreadsheet-token": "sht_test",
		"range":             "sheet1!B2",
		"sheet-id":          "",
		"image":             "./chart.png",
		"name":              "",
		"url":               "",
	}, nil)
	got := mustMarshalSheetsDryRun(t, SheetWriteImage.DryRun(context.Background(), runtime))

	if !strings.Contains(got, `"range":"sheet1!B2:B2"`) {
		t.Fatalf("DryRun range not normalized: %s", got)
	}
	if !strings.Contains(got, `"name":"chart.png"`) {
		t.Fatalf("DryRun name not derived from image path: %s", got)
	}
	// JSON escapes < and > to \u003c and \u003e.
	if !strings.Contains(got, `binary: ./chart.png`) {
		t.Fatalf("DryRun image field not showing binary placeholder: %s", got)
	}
	if !strings.Contains(got, `"description":"JSON upload with inline image bytes"`) {
		t.Fatalf("DryRun description incorrect: %s", got)
	}
}

func TestSheetWriteImageDryRunCustomName(t *testing.T) {
	t.Parallel()
	runtime := newSheetsTestRuntime(t, map[string]string{
		"spreadsheet-token": "sht_test",
		"range":             "sheet1!A1:A1",
		"sheet-id":          "",
		"image":             "./output.png",
		"name":              "revenue_chart.png",
		"url":               "",
	}, nil)
	got := mustMarshalSheetsDryRun(t, SheetWriteImage.DryRun(context.Background(), runtime))

	if !strings.Contains(got, `"name":"revenue_chart.png"`) {
		t.Fatalf("DryRun should use custom name: %s", got)
	}
}

// ── Execute ──────────────────────────────────────────────────────────────────

func TestSheetWriteImageExecuteSendsJSON(t *testing.T) {
	f, stdout, _, reg := cmdutil.TestFactory(t, sheetsTestConfig())

	stub := &httpmock.Stub{
		Method: "POST",
		URL:    "/open-apis/sheets/v2/spreadsheets/shtTOKEN/values_image",
		Body: map[string]interface{}{
			"code": 0,
			"msg":  "success",
			"data": map[string]interface{}{
				"spreadsheetToken": "shtTOKEN",
				"revision":         float64(5),
				"updateRange":      "sheet1!A1:A1",
			},
		},
	}
	reg.Register(stub)

	tmpDir := t.TempDir()
	cmdutil.TestChdir(t, tmpDir)

	// Create a small test image file.
	imgData := []byte{0x89, 0x50, 0x4E, 0x47} // PNG magic bytes
	if err := os.WriteFile("test.png", imgData, 0644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	err := mountAndRunSheets(t, SheetWriteImage, []string{
		"+write-image",
		"--spreadsheet-token", "shtTOKEN",
		"--range", "sheet1!A1:A1",
		"--image", "./test.png",
		"--as", "user",
	}, f, stdout)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the request was sent as JSON (not multipart/form-data).
	if stub.CapturedHeaders == nil {
		t.Fatal("request headers not captured")
	}
	ct := stub.CapturedHeaders.Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Fatalf("Content-Type = %q, want application/json", ct)
	}

	// Verify the captured body contains the image as base64 in JSON.
	var body map[string]interface{}
	if err := json.Unmarshal(stub.CapturedBody, &body); err != nil {
		t.Fatalf("request body is not valid JSON: %v", err)
	}
	if body["range"] != "sheet1!A1:A1" {
		t.Fatalf("body range = %v, want sheet1!A1:A1", body["range"])
	}
	if body["name"] != "test.png" {
		t.Fatalf("body name = %v, want test.png", body["name"])
	}
	if body["image"] == nil {
		t.Fatal("body image field is nil")
	}

	// Verify output contains expected fields.
	if !strings.Contains(stdout.String(), "spreadsheetToken") {
		t.Fatalf("stdout missing spreadsheetToken: %s", stdout.String())
	}
}

func TestSheetWriteImageExecuteRejectsNonexistentFile(t *testing.T) {
	f, _, _, _ := cmdutil.TestFactory(t, sheetsTestConfig())

	tmpDir := t.TempDir()
	cmdutil.TestChdir(t, tmpDir)

	err := mountAndRunSheets(t, SheetWriteImage, []string{
		"+write-image",
		"--spreadsheet-token", "shtTOKEN",
		"--range", "sheet1!A1:A1",
		"--image", "./nonexistent.png",
		"--as", "user",
	}, f, nil)
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}
