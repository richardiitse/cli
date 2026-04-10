// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package slides

import (
	"context"
	"fmt"
	"strings"

	"github.com/larksuite/cli/internal/output"
	"github.com/larksuite/cli/shortcuts/common"
)

const (
	defaultPresentationWidth  = 960
	defaultPresentationHeight = 540
)

// SlidesCreate creates a new Lark Slides presentation with bot auto-grant.
var SlidesCreate = common.Shortcut{
	Service:     "slides",
	Command:     "+create",
	Description: "Create a Lark Slides presentation",
	Risk:        "write",
	AuthTypes:   []string{"user", "bot"},
	Scopes:      []string{"slides:presentation:create"},
	Flags: []common.Flag{
		{Name: "title", Desc: "presentation title"},
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		title := runtime.Str("title")
		dry := common.NewDryRunAPI().
			Desc("Create empty presentation").
			POST("/open-apis/slides_ai/v1/xml_presentations").
			Body(map[string]interface{}{"xml_presentation": map[string]interface{}{"content": buildPresentationXML(title)}})
		if runtime.IsBot() {
			dry.Desc("After creation succeeds in bot mode, the CLI will also try to grant the current CLI user full_access (可管理权限) on the new presentation.")
		}
		return dry
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		title := runtime.Str("title")
		content := buildPresentationXML(title)

		data, err := runtime.CallAPI(
			"POST",
			"/open-apis/slides_ai/v1/xml_presentations",
			nil,
			map[string]interface{}{
				"xml_presentation": map[string]interface{}{
					"content": content,
				},
			},
		)
		if err != nil {
			return err
		}

		presentationID := common.GetString(data, "xml_presentation_id")
		if presentationID == "" {
			return output.Errorf(output.ExitAPI, "api_error", "slides create returned no xml_presentation_id")
		}

		result := map[string]interface{}{
			"xml_presentation_id": presentationID,
			"title":               title,
		}
		if revisionID := common.GetFloat(data, "revision_id"); revisionID > 0 {
			result["revision_id"] = int(revisionID)
		}

		if grant := common.AutoGrantCurrentUserDrivePermission(runtime, presentationID, "slides"); grant != nil {
			result["permission_grant"] = grant
		}

		runtime.Out(result, nil)
		return nil
	},
}

// buildPresentationXML builds the minimal XML for a new empty presentation.
func buildPresentationXML(title string) string {
	escapedTitle := xmlEscape(title)
	if escapedTitle == "" {
		escapedTitle = "Untitled"
	}
	return fmt.Sprintf(
		`<presentation xmlns="http://www.larkoffice.com/sml/2.0" width="%d" height="%d"><title>%s</title></presentation>`,
		defaultPresentationWidth, defaultPresentationHeight, escapedTitle,
	)
}

// xmlEscape escapes special XML characters in text content.
func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}
