// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package slides

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/larksuite/cli/internal/cmdutil"
	"github.com/larksuite/cli/internal/core"
	"github.com/larksuite/cli/internal/httpmock"
	"github.com/larksuite/cli/shortcuts/common"
)

func TestSlidesCreateBasic(t *testing.T) {
	t.Parallel()

	f, stdout, _, reg := cmdutil.TestFactory(t, slidesTestConfig(t, ""))
	reg.Register(&httpmock.Stub{
		Method: "POST",
		URL:    "/open-apis/slides_ai/v1/xml_presentations",
		Body: map[string]interface{}{
			"code": 0,
			"msg":  "ok",
			"data": map[string]interface{}{
				"xml_presentation_id": "pres_abc123",
				"revision_id":         1,
			},
		},
	})

	err := runSlidesCreateShortcut(t, f, stdout, []string{
		"+create",
		"--title", "项目汇报",
		"--as", "user",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data := decodeSlidesCreateEnvelope(t, stdout)
	if data["xml_presentation_id"] != "pres_abc123" {
		t.Fatalf("xml_presentation_id = %v, want pres_abc123", data["xml_presentation_id"])
	}
	if data["title"] != "项目汇报" {
		t.Fatalf("title = %v, want 项目汇报", data["title"])
	}
	if _, ok := data["permission_grant"]; ok {
		t.Fatalf("did not expect permission_grant in user mode")
	}
}

func TestSlidesCreateBotAutoGrant(t *testing.T) {
	t.Parallel()

	f, stdout, _, reg := cmdutil.TestFactory(t, slidesTestConfig(t, "ou_current_user"))
	reg.Register(&httpmock.Stub{
		Method: "POST",
		URL:    "/open-apis/slides_ai/v1/xml_presentations",
		Body: map[string]interface{}{
			"code": 0,
			"msg":  "ok",
			"data": map[string]interface{}{
				"xml_presentation_id": "pres_bot",
				"revision_id":         1,
			},
		},
	})
	reg.Register(&httpmock.Stub{
		Method: "POST",
		URL:    "/open-apis/drive/v1/permissions/pres_bot/members",
		Body: map[string]interface{}{
			"code": 0,
			"msg":  "ok",
			"data": map[string]interface{}{
				"member": map[string]interface{}{
					"member_id":   "ou_current_user",
					"member_type": "openid",
					"perm":        "full_access",
				},
			},
		},
	})

	err := runSlidesCreateShortcut(t, f, stdout, []string{
		"+create",
		"--title", "Bot PPT",
		"--as", "bot",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data := decodeSlidesCreateEnvelope(t, stdout)
	grant, _ := data["permission_grant"].(map[string]interface{})
	if grant["status"] != common.PermissionGrantGranted {
		t.Fatalf("permission_grant.status = %v, want %q", grant["status"], common.PermissionGrantGranted)
	}
	if !strings.Contains(grant["message"].(string), "presentation") {
		t.Fatalf("permission_grant.message = %q, want 'presentation' mention", grant["message"])
	}
}

func TestSlidesCreateBotSkippedWithoutCurrentUser(t *testing.T) {
	t.Parallel()

	f, stdout, _, reg := cmdutil.TestFactory(t, slidesTestConfig(t, ""))
	reg.Register(&httpmock.Stub{
		Method: "POST",
		URL:    "/open-apis/slides_ai/v1/xml_presentations",
		Body: map[string]interface{}{
			"code": 0,
			"msg":  "ok",
			"data": map[string]interface{}{
				"xml_presentation_id": "pres_no_user",
				"revision_id":         1,
			},
		},
	})

	err := runSlidesCreateShortcut(t, f, stdout, []string{
		"+create",
		"--title", "No User PPT",
		"--as", "bot",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data := decodeSlidesCreateEnvelope(t, stdout)
	grant, _ := data["permission_grant"].(map[string]interface{})
	if grant["status"] != common.PermissionGrantSkipped {
		t.Fatalf("permission_grant.status = %v, want %q", grant["status"], common.PermissionGrantSkipped)
	}
}

func TestXmlEscape(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input, want string
	}{
		{"hello", "hello"},
		{"a&b", "a&amp;b"},
		{"<script>", "&lt;script&gt;"},
		{`"quoted"`, "&quot;quoted&quot;"},
		{"it's", "it&apos;s"},
	}
	for _, tt := range tests {
		got := xmlEscape(tt.input)
		if got != tt.want {
			t.Errorf("xmlEscape(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// ── helpers ──

func slidesTestConfig(t *testing.T, userOpenID string) *core.CliConfig {
	t.Helper()
	replacer := strings.NewReplacer("/", "-", " ", "-")
	suffix := replacer.Replace(strings.ToLower(t.Name()))
	return &core.CliConfig{
		AppID:      "test-slides-create-" + suffix,
		AppSecret:  "secret-slides-create-" + suffix,
		Brand:      core.BrandFeishu,
		UserOpenId: userOpenID,
	}
}

func runSlidesCreateShortcut(t *testing.T, f *cmdutil.Factory, stdout *bytes.Buffer, args []string) error {
	t.Helper()
	parent := &cobra.Command{Use: "slides"}
	SlidesCreate.Mount(parent, f)
	parent.SetArgs(args)
	parent.SilenceErrors = true
	parent.SilenceUsage = true
	if stdout != nil {
		stdout.Reset()
	}
	return parent.Execute()
}

func decodeSlidesCreateEnvelope(t *testing.T, stdout *bytes.Buffer) map[string]interface{} {
	t.Helper()
	var envelope map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil {
		t.Fatalf("failed to decode output: %v\nraw=%s", err, stdout.String())
	}
	data, _ := envelope["data"].(map[string]interface{})
	if data == nil {
		t.Fatalf("missing data in output envelope: %#v", envelope)
	}
	return data
}
