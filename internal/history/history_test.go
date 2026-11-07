package history

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ankitaniket/clean-me/internal/models"
)

func TestRecordListRoundTripNewestFirst(t *testing.T) {
	base := t.TempDir()

	t1 := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 2, 2, 11, 0, 0, 0, time.UTC)

	out1 := []models.ItemOutcome{
		{Item: models.Item{Stage: "node", Path: "/proj/node_modules"}, Reclaimed: 100, TrashedTo: "/trash/a"},
	}
	out2 := []models.ItemOutcome{
		{Item: models.Item{Stage: "python", Path: "/proj/.venv"}, Reclaimed: 250, TrashedTo: "/trash/b"},
		{Item: models.Item{Stage: "cache", Path: "/proj/cache"}, Reclaimed: 50}, // freed but not trashed
	}

	if err := Record(base, t1, out1); err != nil {
		t.Fatalf("Record #1: %v", err)
	}
	if err := Record(base, t2, out2); err != nil {
		t.Fatalf("Record #2: %v", err)
	}

	runs, err := List(base)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(runs) != 2 {
		t.Fatalf("got %d runs, want 2", len(runs))
	}

	// Newest-first: t2 before t1.
	if runs[0].Time != t2.Format(time.RFC3339) {
		t.Errorf("runs[0].Time = %q, want %q", runs[0].Time, t2.Format(time.RFC3339))
	}
	if runs[1].Time != t1.Format(time.RFC3339) {
		t.Errorf("runs[1].Time = %q, want %q", runs[1].Time, t1.Format(time.RFC3339))
	}

	// Newest run freed 250+50, with only the trashed item restorable.
	if runs[0].Freed != 300 {
		t.Errorf("runs[0].Freed = %d, want 300", runs[0].Freed)
	}
	if len(runs[0].Items) != 1 {
		t.Fatalf("runs[0] has %d restorable items, want 1", len(runs[0].Items))
	}
	got := runs[0].Items[0]
	if got.Stage != "python" || got.OriginalPath != "/proj/.venv" || got.TrashedTo != "/trash/b" || got.Reclaimed != 250 {
		t.Errorf("runs[0].Items[0] = %+v", got)
	}

	if runs[1].Freed != 100 || len(runs[1].Items) != 1 {
		t.Errorf("runs[1] = %+v", runs[1])
	}
}

func TestRecordWritesNothingWhenEmpty(t *testing.T) {
	base := t.TempDir()
	now := time.Date(2026, 3, 3, 12, 0, 0, 0, time.UTC)

	outcomes := []models.ItemOutcome{
		// Failed outcome: not successful, not trashed.
		{Item: models.Item{Stage: "node", Path: "/x"}, Err: os.ErrPermission, TrashedTo: "/trash/x"},
		// Report-only / nothing freed, nothing trashed.
		{Item: models.Item{Stage: "git", Path: "/y"}},
	}

	if err := Record(base, now, outcomes); err != nil {
		t.Fatalf("Record: %v", err)
	}

	if _, err := os.Stat(logPath(base)); !os.IsNotExist(err) {
		t.Fatalf("expected no log file, stat err = %v", err)
	}

	runs, err := List(base)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(runs) != 0 {
		t.Fatalf("got %d runs, want 0", len(runs))
	}
}

func TestRestoreMovesFileBack(t *testing.T) {
	dir := t.TempDir()
	trashed := filepath.Join(dir, "trash", "item")
	if err := os.MkdirAll(filepath.Dir(trashed), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(trashed, []byte("payload"), 0600); err != nil {
		t.Fatal(err)
	}

	// Original path does not exist yet and lives under a missing parent dir.
	original := filepath.Join(dir, "restored", "nested", "item")

	item := RestoredItem{Stage: "node", OriginalPath: original, TrashedTo: trashed, Reclaimed: 7}
	if err := Restore(item); err != nil {
		t.Fatalf("Restore: %v", err)
	}

	if _, err := os.Stat(original); err != nil {
		t.Fatalf("original path not restored: %v", err)
	}
	if _, err := os.Stat(trashed); !os.IsNotExist(err) {
		t.Fatalf("trashed path should be gone, stat err = %v", err)
	}
	data, err := os.ReadFile(original)
	if err != nil || string(data) != "payload" {
		t.Fatalf("restored content = %q, err = %v", data, err)
	}
}

func TestRestoreRefusesWhenOriginalExists(t *testing.T) {
	dir := t.TempDir()
	trashed := filepath.Join(dir, "trash-item")
	original := filepath.Join(dir, "orig-item")
	if err := os.WriteFile(trashed, []byte("new"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(original, []byte("existing"), 0600); err != nil {
		t.Fatal(err)
	}

	item := RestoredItem{OriginalPath: original, TrashedTo: trashed}
	if err := Restore(item); err == nil {
		t.Fatal("Restore: expected error when original exists, got nil")
	}

	// The existing original must be untouched, and the trashed file preserved.
	if data, _ := os.ReadFile(original); string(data) != "existing" {
		t.Errorf("original was clobbered: %q", data)
	}
	if _, err := os.Stat(trashed); err != nil {
		t.Errorf("trashed file should still exist: %v", err)
	}
}

func TestRestoreGuards(t *testing.T) {
	if err := Restore(RestoredItem{OriginalPath: "/dst"}); err == nil {
		t.Error("expected error for empty TrashedTo")
	}
	if err := Restore(RestoredItem{TrashedTo: "/does/not/exist/anywhere", OriginalPath: "/dst"}); err == nil {
		t.Error("expected error for missing trashed path")
	}
}

func TestDirDefault(t *testing.T) {
	got := Dir("/base")
	want := filepath.Join("/base", ".config", "clean-me", "history")
	if got != want {
		t.Errorf("Dir(/base) = %q, want %q", got, want)
	}
}
