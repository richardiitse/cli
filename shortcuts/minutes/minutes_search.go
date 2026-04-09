// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package minutes

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/larksuite/cli/internal/output"
	"github.com/larksuite/cli/shortcuts/common"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
)

const (
	defaultMinutesSearchPageSize = 15
	maxMinutesSearchPageSize     = 200
	maxMinutesSearchQueryLen     = 50
)

func parseTimeRange(runtime *common.RuntimeContext) (string, string, error) {
	start := strings.TrimSpace(runtime.Str("start"))
	end := strings.TrimSpace(runtime.Str("end"))
	if start == "" && end == "" {
		return "", "", nil
	}

	var startTime, endTime string
	if start != "" {
		parsed, err := toRFC3339(start)
		if err != nil {
			return "", "", output.ErrValidation("--start: %v", err)
		}
		startTime = parsed
	}
	if end != "" {
		parsed, err := toRFC3339(end, "end")
		if err != nil {
			return "", "", output.ErrValidation("--end: %v", err)
		}
		endTime = parsed
	}
	if startTime != "" && endTime != "" {
		st, _ := time.Parse(time.RFC3339, startTime)
		et, _ := time.Parse(time.RFC3339, endTime)
		if st.After(et) {
			return "", "", output.ErrValidation("--start (%s) is after --end (%s)", start, end)
		}
	}
	return startTime, endTime, nil
}

func toRFC3339(input string, hint ...string) (string, error) {
	ts, err := common.ParseTime(input, hint...)
	if err != nil {
		return "", err
	}
	sec, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return "", fmt.Errorf("invalid timestamp %q: %w", ts, err)
	}
	return time.Unix(sec, 0).Format(time.RFC3339), nil
}

func resolveUserIDs(ids []string, runtime *common.RuntimeContext) []string {
	if len(ids) == 0 {
		return nil
	}
	currentUserID := runtime.UserOpenId()
	seen := make(map[string]struct{}, len(ids))
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if strings.EqualFold(id, "me") && currentUserID != "" {
			id = currentUserID
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func buildTimeFilter(startTime, endTime string) map[string]interface{} {
	if startTime == "" && endTime == "" {
		return nil
	}
	timeRange := map[string]interface{}{}
	if startTime != "" {
		timeRange["start_time"] = startTime
	}
	if endTime != "" {
		timeRange["end_time"] = endTime
	}
	return timeRange
}

func buildMinutesSearchFilter(runtime *common.RuntimeContext, startTime, endTime string) map[string]interface{} {
	filter := map[string]interface{}{}

	ownerIDs := resolveUserIDs(common.SplitCSV(runtime.Str("owner-ids")), runtime)
	if len(ownerIDs) > 0 {
		filter["owner_ids"] = ownerIDs
	}

	participantIDs := resolveUserIDs(common.SplitCSV(runtime.Str("participant-ids")), runtime)
	if len(participantIDs) > 0 {
		filter["participant_ids"] = participantIDs
	}

	if timeRange := buildTimeFilter(startTime, endTime); timeRange != nil {
		filter["create_time"] = timeRange
	}

	if len(filter) == 0 {
		return nil
	}
	return filter
}

func buildMinutesSearchBody(runtime *common.RuntimeContext, startTime, endTime string) map[string]interface{} {
	body := map[string]interface{}{}

	if q := strings.TrimSpace(runtime.Str("query")); q != "" {
		body["query"] = q
	}

	if filter := buildMinutesSearchFilter(runtime, startTime, endTime); filter != nil {
		body["filter"] = filter
	}

	return body
}

func buildMinutesSearchParams(runtime *common.RuntimeContext) larkcore.QueryParams {
	params := larkcore.QueryParams{}

	pageSize := strings.TrimSpace(runtime.Str("page-size"))
	if pageSize == "" {
		pageSize = fmt.Sprintf("%d", defaultMinutesSearchPageSize)
	}
	params["page_size"] = []string{pageSize}

	if pageToken := strings.TrimSpace(runtime.Str("page-token")); pageToken != "" {
		params["page_token"] = []string{pageToken}
	}

	return params
}

func minuteSearchItems(data map[string]interface{}) []interface{} {
	return common.GetSlice(data, "items")
}

func minuteSearchToken(item map[string]interface{}) string {
	return common.GetString(item, "token")
}

func minuteSearchDisplayInfo(item map[string]interface{}) string {
	return common.GetString(item, "display_info")
}

func minuteSearchDescription(item map[string]interface{}) string {
	meta := common.GetMap(item, "meta_data")
	return common.GetString(meta, "description")
}

func minuteSearchAppLink(item map[string]interface{}) string {
	meta := common.GetMap(item, "meta_data")
	return common.GetString(meta, "app_link")
}

func minuteSearchAvatar(item map[string]interface{}) string {
	meta := common.GetMap(item, "meta_data")
	return common.GetString(meta, "avatar")
}

var MinutesSearch = common.Shortcut{
	Service:     "minutes",
	Command:     "+search",
	Description: "Search minutes by keyword, owners, participants, and time range",
	Risk:        "read",
	Scopes:      []string{"minutes:minutes.search:read"},
	AuthTypes:   []string{"user"},
	HasFormat:   true,
	Flags: []common.Flag{
		{Name: "query", Desc: "search keyword"},
		{Name: "owner-ids", Desc: "owner open_id list, comma-separated (use \"me\" for current user)"},
		{Name: "participant-ids", Desc: "participant open_id list, comma-separated (use \"me\" for current user)"},
		{Name: "start", Desc: "time lower bound (ISO 8601 or YYYY-MM-DD)"},
		{Name: "end", Desc: "time upper bound (ISO 8601 or YYYY-MM-DD)"},
		{Name: "page-token", Desc: "page token for next page"},
		{Name: "page-size", Default: "15", Desc: "page size, 1-200 (default 15)"},
	},
	Validate: func(ctx context.Context, runtime *common.RuntimeContext) error {
		if _, _, err := parseTimeRange(runtime); err != nil {
			return err
		}
		if q := strings.TrimSpace(runtime.Str("query")); q != "" && utf8.RuneCountInString(q) > maxMinutesSearchQueryLen {
			return output.ErrValidation("--query: length must be between 1 and 50 characters")
		}
		if _, err := common.ValidatePageSize(runtime, "page-size", defaultMinutesSearchPageSize, 1, maxMinutesSearchPageSize); err != nil {
			return err
		}
		for _, id := range resolveUserIDs(common.SplitCSV(runtime.Str("owner-ids")), runtime) {
			if _, err := common.ValidateUserID(id); err != nil {
				return err
			}
		}
		for _, id := range resolveUserIDs(common.SplitCSV(runtime.Str("participant-ids")), runtime) {
			if _, err := common.ValidateUserID(id); err != nil {
				return err
			}
		}
		for _, flag := range []string{"query", "owner-ids", "participant-ids", "start", "end"} {
			if strings.TrimSpace(runtime.Str(flag)) != "" {
				return nil
			}
		}
		return common.FlagErrorf("specify at least one of --query, --owner-ids, --participant-ids, --start, or --end")
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		startTime, endTime, err := parseTimeRange(runtime)
		if err != nil {
			return common.NewDryRunAPI().Set("error", err.Error())
		}
		params := buildMinutesSearchParams(runtime)
		dryRunParams := map[string]interface{}{}
		for key, values := range params {
			if len(values) == 1 {
				dryRunParams[key] = values[0]
			}
		}
		dryRun := common.NewDryRunAPI().
			POST("/open-apis/minutes/v1/minutes/search")
		if len(dryRunParams) > 0 {
			dryRun.Params(dryRunParams)
		}
		return dryRun.Body(buildMinutesSearchBody(runtime, startTime, endTime))
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		startTime, endTime, err := parseTimeRange(runtime)
		if err != nil {
			return err
		}

		params := map[string]interface{}{}
		pageSize, _ := strconv.Atoi(buildMinutesSearchParams(runtime).Get("page_size"))
		params["page_size"] = pageSize
		if pageToken := strings.TrimSpace(runtime.Str("page-token")); pageToken != "" {
			params["page_token"] = pageToken
		}

		data, err := runtime.CallAPI(http.MethodPost, "/open-apis/minutes/v1/minutes/search", params, buildMinutesSearchBody(runtime, startTime, endTime))
		if err != nil {
			return err
		}
		if data == nil {
			data = map[string]interface{}{}
		}

		items := minuteSearchItems(data)
		hasMore, _ := data["has_more"].(bool)
		pageToken, _ := data["page_token"].(string)

		outData := map[string]interface{}{
			"items":      items,
			"total":      data["total"],
			"has_more":   data["has_more"],
			"page_token": data["page_token"],
		}

		runtime.OutFormat(outData, &output.Meta{Count: len(items)}, func(w io.Writer) {
			if len(items) == 0 {
				fmt.Fprintln(w, "No minutes.")
				return
			}

			var rows []map[string]interface{}
			for _, raw := range items {
				item, _ := raw.(map[string]interface{})
				if item == nil {
					continue
				}
				rows = append(rows, map[string]interface{}{
					"token":        minuteSearchToken(item),
					"display_info": common.TruncateStr(minuteSearchDisplayInfo(item), 40),
					"description":  common.TruncateStr(minuteSearchDescription(item), 40),
					"app_link":     common.TruncateStr(minuteSearchAppLink(item), 80),
					"avatar":       common.TruncateStr(minuteSearchAvatar(item), 80),
				})
			}
			output.PrintTable(w, rows)
			if hasMore {
				fmt.Fprintf(w, "\n(more available, page_token: %s)\n", pageToken)
			}
		})
		return nil
	},
}
