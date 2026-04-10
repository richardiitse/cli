// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package calendar

import (
	"errors"
	"testing"
	"time"

	"github.com/larksuite/cli/internal/output"
)

func TestCollectRoomFindResults_LimitsConcurrency(t *testing.T) {
	slots := []roomFindSlot{
		{Start: "2026-03-27T14:00:00+08:00", End: "2026-03-27T15:00:00+08:00"},
		{Start: "2026-03-27T15:00:00+08:00", End: "2026-03-27T16:00:00+08:00"},
		{Start: "2026-03-27T16:00:00+08:00", End: "2026-03-27T17:00:00+08:00"},
	}

	entered := make(chan struct{}, len(slots))
	release := make(chan struct{})
	done := make(chan *roomFindOutput, 1)
	errCh := make(chan error, 1)

	go func() {
		out, err := collectRoomFindResults(slots, 2, func(slot roomFindSlot) ([]*roomFindSuggestion, error) {
			entered <- struct{}{}
			<-release
			return []*roomFindSuggestion{{RoomName: slot.Start}}, nil
		})
		errCh <- err
		done <- out
	}()

	for range 2 {
		select {
		case <-entered:
		case <-time.After(200 * time.Millisecond):
			t.Fatal("timed out waiting for room-find workers to start")
		}
	}

	select {
	case <-entered:
		t.Fatal("room-find exceeded the configured concurrency limit")
	case <-time.After(50 * time.Millisecond):
	}

	close(release)

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("collectRoomFindResults returned error: %v", err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for room-find results")
	}

	out := <-done
	if len(out.TimeSlots) != len(slots) {
		t.Fatalf("expected %d time slots, got %d", len(slots), len(out.TimeSlots))
	}
	if len(out.PartialFailures) != 0 {
		t.Fatalf("expected no partial failures, got %d", len(out.PartialFailures))
	}
}

func TestCollectRoomFindResults_ReturnsErrorWhenAllSlotsFail(t *testing.T) {
	slots := []roomFindSlot{
		{Start: "2026-03-27T14:00:00+08:00", End: "2026-03-27T15:00:00+08:00"},
		{Start: "2026-03-27T15:00:00+08:00", End: "2026-03-27T16:00:00+08:00"},
	}

	wantErr := output.ErrNetwork("backend unavailable")

	out, err := collectRoomFindResults(slots, 2, func(roomFindSlot) ([]*roomFindSuggestion, error) {
		return nil, wantErr
	})
	if err == nil {
		t.Fatal("expected error when all room-find slots fail")
	}

	var exitErr *output.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected structured exit error, got %T", err)
	}
	if exitErr.Code != output.ExitNetwork {
		t.Fatalf("expected exit code %d, got %d", output.ExitNetwork, exitErr.Code)
	}
	if out == nil {
		t.Fatal("expected output details for total failure")
	}
	if got := len(out.PartialFailures); got != len(slots) {
		t.Fatalf("expected %d partial failures, got %d", len(slots), got)
	}
	if len(out.TimeSlots) != 0 {
		t.Fatalf("expected no successful time slots, got %d", len(out.TimeSlots))
	}
}
