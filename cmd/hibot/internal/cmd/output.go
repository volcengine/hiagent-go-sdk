package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"gopkg.in/yaml.v3"
)

// renderJSON pretty-prints v as JSON (indented).
func renderJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// renderYAML renders v as YAML.
func renderYAML(w io.Writer, v any) error {
	// Round-trip through JSON to keep struct field name capitalization
	// (the SDK types use uppercase JSON tags) consistent.
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	defer enc.Close()
	return enc.Encode(raw)
}

// renderTable writes a header + rows table using text/tabwriter.
func renderTable(w io.Writer, headers []string, rows [][]string) error {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	if len(headers) > 0 {
		_, _ = fmt.Fprintln(tw, strings.Join(headers, "\t"))
	}
	for _, row := range rows {
		_, _ = fmt.Fprintln(tw, strings.Join(row, "\t"))
	}
	return tw.Flush()
}

// emit picks an output renderer based on format. table requires headers/rows;
// json/yaml fall back to a structured representation passed via tableRow == nil.
type emitter struct {
	format string
	w      io.Writer
}

func newEmitter(format string, w io.Writer) emitter {
	return emitter{format: format, w: w}
}

// emitObject renders an object: JSON/YAML uses obj as-is; table renders rows.
func (e emitter) emitObject(obj any, headers []string, rows [][]string) error {
	switch e.format {
	case "json":
		return renderJSON(e.w, obj)
	case "yaml":
		return renderYAML(e.w, obj)
	default:
		return renderTable(e.w, headers, rows)
	}
}

// emitMessage prints a status / informational line. Suppressed when output is
// json/yaml so machine consumers don't get pollution on stdout.
func (e emitter) emitMessage(format string, args ...any) {
	if e.format == "json" || e.format == "yaml" {
		return
	}
	fmt.Fprintf(e.w, format+"\n", args...)
}

// truncate cuts s to at most n runes (for table cells).
func truncate(s string, n int) string {
	if n <= 0 {
		return s
	}
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}
