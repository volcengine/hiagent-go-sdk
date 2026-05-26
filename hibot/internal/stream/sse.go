package stream

import (
	"bufio"
	"io"
	"strings"
)

// Frame is a single Server-Sent Events frame.
type Frame struct {
	Event string
	Data  string
}

// Decoder reads text/event-stream frames.
type Decoder struct {
	reader *bufio.Reader
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{reader: bufio.NewReader(r)}
}

func (d *Decoder) Next() (Frame, error) {
	var frame Frame
	var data []string
	for {
		line, err := d.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF && (frame.Event != "" || len(data) > 0) {
				frame.Data = strings.Join(data, "\n")
				return frame, nil
			}
			return Frame{}, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			if frame.Event == "" && len(data) == 0 {
				continue
			}
			frame.Data = strings.Join(data, "\n")
			return frame, nil
		}
		if strings.HasPrefix(line, ":") {
			continue
		}
		if value, ok := strings.CutPrefix(line, "event:"); ok {
			frame.Event = strings.TrimSpace(value)
			continue
		}
		if value, ok := strings.CutPrefix(line, "data:"); ok {
			data = append(data, strings.TrimSpace(value))
		}
	}
}
