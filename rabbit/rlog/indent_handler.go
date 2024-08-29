package rlog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"sync"
	"time"
)

type IndentHandler struct {
	// slog handler
	opts slog.HandlerOptions
	mu   *sync.Mutex
	out  io.Writer
}

func NewIndentHandler(w io.Writer, opts slog.HandlerOptions) *IndentHandler {
	h := IndentHandler{
		opts: opts,
		mu:   &sync.Mutex{},
		out:  w,
	}

	if h.opts.Level == nil {
		h.opts.Level = slog.LevelInfo
	}
	return &h
}

/* method that implement slog.Handler */

var _ slog.Handler = (*IndentHandler)(nil)

// Enabled implements slog.Handler.
func (e *IndentHandler) Enabled(_ context.Context, l slog.Level) bool {
	return l >= e.opts.Level.Level()
}

// Handle implements slog.Handler.
func (e *IndentHandler) Handle(ctx context.Context, rec slog.Record) error {

	var err error
	buf := make([]byte, 0, 1024)

	// TIME
	if !rec.Time.IsZero() {
		buf = e.appendAttr(buf, slog.Time(slog.TimeKey, rec.Time), 0)
	}
	// LEVEL
	buf = e.appendAttr(buf, slog.Any(slog.LevelKey, rec.Level), 0)

	// SOURCE
	if rec.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{rec.PC})
		f, _ := fs.Next()
		buf = e.appendAttr(buf, slog.String(slog.SourceKey, fmt.Sprintf("%s:%d", f.File, f.Line)), 0)
	}

	// MESSAGE
	buf = e.appendAttr(buf, slog.String(slog.MessageKey, rec.Message), 0)
	/*
		TODO: output the Attrs and groups from WithAttrs and WithGroup.
	*/

	// ARGS
	var indentLevel = 0
	rec.Attrs(func(a slog.Attr) bool {
		buf = e.appendAttr(buf, a, indentLevel)
		return true
	})

	buf = append(buf, "---\n"...)
	e.mu.Lock()
	defer e.mu.Unlock()
	_, err = e.out.Write(buf)
	return err
}

func (e *IndentHandler) appendAttr(buf []byte, a slog.Attr, indentLevel int) []byte {
	// Resolve the Attr's value before doing anything else.
	a.Value = a.Value.Resolve()
	// Ignore empty Attrs.
	if a.Equal(slog.Attr{}) {
		return buf
	}

	// Indent 4 spaces per level.
	buf = fmt.Appendf(buf, "%*s", indentLevel*4, "")
	switch a.Value.Kind() {
	case slog.KindString:
		// Quote string values, to make them easy to parse.
		buf = fmt.Appendf(buf, "%s: %q\n", a.Key, a.Value.String())
	case slog.KindTime:
		// Write times in a standard way, without the monotonic time.
		buf = fmt.Appendf(buf, "%s: %s\n", a.Key, a.Value.Time().Format(time.RFC3339Nano))
	case slog.KindGroup:
		attrs := a.Value.Group()
		// Ignore empty groups.
		if len(attrs) == 0 {
			return buf
		}
		// If the key is non-empty, write it out and indent the rest of the attrs.
		// Otherwise, inline the attrs.
		if a.Key != "" {
			buf = fmt.Appendf(buf, "%s:\n", a.Key)
			indentLevel++
		}
		for _, attr := range attrs {
			buf = e.appendAttr(buf, attr, indentLevel)
		}

	default:
		buf = fmt.Appendf(buf, "%s: %s\n", a.Key, a.Value)
	}
	return buf

}

// WithAttrs implements slog.Handler.
func (e *IndentHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	panic("unimplemented")
}

// WithGroup implements slog.Handler.
func (e *IndentHandler) WithGroup(name string) slog.Handler {
	panic("unimplemented")
}
