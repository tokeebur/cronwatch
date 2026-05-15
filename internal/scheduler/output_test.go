package scheduler

import (
	"strings"
	"testing"
)

func TestProcessOutput_NoTruncation(t *testing.T) {
	input := "hello world"
	got := processOutput(input, 0, false)
	if got != input {
		t.Fatalf("expected %q, got %q", input, got)
	}
}

func TestProcessOutput_TruncatesLongOutput(t *testing.T) {
	input := strings.Repeat("a", 8192)
	got := processOutput(input, 4096, false)
	if len(got) != 4096 {
		t.Fatalf("expected length 4096, got %d", len(got))
	}
}

func TestProcessOutput_ShortOutputNotTruncated(t *testing.T) {
	input := "short"
	got := processOutput(input, 4096, false)
	if got != input {
		t.Fatalf("expected %q, got %q", input, got)
	}
}

func TestStripANSICodes_RemovesEscapes(t *testing.T) {
	input := "\x1b[31mred text\x1b[0m"
	got := stripANSICodes(input)
	want := "red text"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestStripANSICodes_PlainTextUnchanged(t *testing.T) {
	input := "plain text"
	got := stripANSICodes(input)
	if got != input {
		t.Fatalf("expected %q, got %q", input, got)
	}
}

func TestStripANSICodes_MultipleSequences(t *testing.T) {
	input := "\x1b[1;32mbold green\x1b[0m normal"
	got := stripANSICodes(input)
	want := "bold green normal"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestProcessOutput_StripAndTruncate(t *testing.T) {
	base := strings.Repeat("x", 6000)
	input := "\x1b[31m" + base + "\x1b[0m"
	got := processOutput(input, 4096, true)
	if strings.ContainsRune(got, '\x1b') {
		t.Fatal("expected ANSI codes to be stripped")
	}
	if len(got) > 4096 {
		t.Fatalf("expected length <= 4096, got %d", len(got))
	}
}

func TestOutputFromJob_Defaults(t *testing.T) {
	job := map[string]interface{}{}
	maxBytes, stripANSI := OutputFromJob(job)
	if maxBytes != defaultMaxBytes {
		t.Fatalf("expected default maxBytes %d, got %d", defaultMaxBytes, maxBytes)
	}
	if stripANSI {
		t.Fatal("expected stripANSI false by default")
	}
}

func TestOutputFromJob_CustomValues(t *testing.T) {
	job := map[string]interface{}{
		"output_max_bytes":  1024,
		"output_strip_ansi": true,
	}
	maxBytes, stripANSI := OutputFromJob(job)
	if maxBytes != 1024 {
		t.Fatalf("expected 1024, got %d", maxBytes)
	}
	if !stripANSI {
		t.Fatal("expected stripANSI true")
	}
}

func TestOutputFromJob_ZeroMaxBytesAllowed(t *testing.T) {
	job := map[string]interface{}{"output_max_bytes": 0}
	maxBytes, _ := OutputFromJob(job)
	if maxBytes != 0 {
		t.Fatalf("expected 0, got %d", maxBytes)
	}
}
