package rlog

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/odit-bit/messenger/rabbit/rlog/internal/bpool"
)

var _ slog.Handler = (*JSONHandler)(nil)

type Publisher interface {
	Publish(ctx context.Context, body []byte, key string) error
}

type JSONHandler struct {
	// slog handler
	groups []groupAttr
	opts   slog.HandlerOptions
	mu     *sync.Mutex
	out    io.Writer
	e      Publisher
}

func NewJSONHandler(pub Publisher, opts slog.HandlerOptions) *JSONHandler {
	if pub == nil {
		panicOnError(fmt.Errorf("publisher cannot be nil"), "failed init handler")
	}
	h := JSONHandler{
		opts: opts,
		mu:   &sync.Mutex{},
		out:  nil,
		e:    pub,
	}

	if h.opts.Level == nil {
		h.opts.Level = slog.LevelInfo
	}
	return &h
}

func (e *JSONHandler) WithWriter(w io.Writer) {
	e.out = w
}

// Enabled implements slog.Handler.
func (e *JSONHandler) Enabled(ctx context.Context, l slog.Level) bool {
	return l >= e.opts.Level.Level()
}

type groupAttr struct {
	name       string
	attributes []slog.Attr
}

func (e *JSONHandler) withGroupAttr(groups groupAttr) slog.Handler {
	e2 := *e
	e2.groups = make([]groupAttr, len(e.groups)+1)
	copy(e2.groups, e.groups)
	e2.groups[len(e2.groups)-1] = groups
	return &e2
}

// WithAttrs implements slog.Handler.
func (e *JSONHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return e
	}
	return e.withGroupAttr(groupAttr{attributes: attrs})
}

// WithGroup implements slog.Handler.
func (e *JSONHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return e
	}

	return e.withGroupAttr(groupAttr{name: name})
}

// Handle implements slog.Handler.
func (h *JSONHandler) Handle(ctx context.Context, rec slog.Record) error {

	var err error

	state := getJsonBuilder()
	defer state.free()
	state.buf.WriteByte('{')
	// TIME
	if !rec.Time.IsZero() {
		state.appendKey(slog.TimeKey)
		state.appendTime(rec.Time.Round(0))
	}

	// LEVEL
	state.appendAttr(slog.String(slog.LevelKey, rec.Level.String()))

	// SOURCE
	if rec.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{rec.PC})
		f, _ := fs.Next()
		state.appendAttr(slog.String(slog.SourceKey, fmt.Sprintf("%s:%d", f.File, f.Line)))
	}

	// MESSAGE
	state.appendAttr(slog.String(slog.MessageKey, rec.Message))
	// /*
	// 	TODO: output the Attrs and groups from WithAttrs and WithGroup.
	groups := h.groups
	if rec.NumAttrs() == 0 {
		for len(groups) > 0 && groups[len(groups)-1].name != "" {
			groups = groups[:len(groups)-1]
		}
	}

	for _, g := range groups {
		if g.name != "" {
			//
		} else {
			for _, a := range g.attributes {
				state.appendAttr(a)
			}
		}
	}
	// */

	// ATTR
	rec.Attrs(func(a slog.Attr) bool {
		state.appendAttr(a)
		return true
	})

	state.buf.WriteByte('}')
	state.buf.WriteByte('\n')

	// publish
	h.mu.Lock()
	defer h.mu.Unlock()

	msg := *state.buf
	keyRoute := rec.Level.String()
	if h.e != nil {
		if err := h.e.Publish(ctx, msg, keyRoute); err != nil {
			return err
		}
	}
	if h.out != nil {
		if _, err = h.out.Write(msg); err != nil {
			return err
		}
	}
	return nil
}

type jsonBuilder struct {
	buf *bpool.B
}

func getJsonBuilder() jsonBuilder {
	return jsonBuilder{
		buf: bpool.New(),
	}
}

func (jb *jsonBuilder) free() {
	jb.buf.Free()
}

func (jb *jsonBuilder) appendAttr(a slog.Attr) {
	// Resolve the Attr's value before doing anything else.
	a.Value = a.Value.Resolve()
	// Ignore empty Attrs.
	if a.Equal(slog.Attr{}) {
		return
	}

	// KEY
	jb.buf.WriteByte(',')
	jb.appendKey(a.Key)

	//VALUE
	switch a.Value.Kind() {

	case slog.KindString:
		jb.appendString(a.Value.String())
		return

	case slog.KindTime:
		// jb.buf.WriteString(fmt.Sprintf("%q:%q", a.Key, a.Value.Time().Format(time.RFC3339Nano)))
		jb.appendTime(a.Value.Time())

	case slog.KindInt64:
		*jb.buf = strconv.AppendInt(*jb.buf, a.Value.Int64(), 10)

	case slog.KindBool:
		*jb.buf = strconv.AppendBool(*jb.buf, a.Value.Bool())

	case slog.KindDuration:
		// Do what json.Marshal does.
		*jb.buf = strconv.AppendInt(*jb.buf, int64(a.Value.Duration()), 10)

	case slog.KindAny:
		e := a.Value.Any()
		_, isMarshaler := e.(*json.Marshaler)
		if err, ok := e.(error); ok && !isMarshaler {
			jb.appendString(err.Error())
		} else {
			jb.appendJSON(e)
		}

	default:
		fmt.Printf("unimplemented kind: %v \n", a.Value.Kind())
	}
}

func (jb *jsonBuilder) appendTime(a time.Time) {
	jb.buf.WriteByte('"')
	*jb.buf = a.AppendFormat(*jb.buf, time.RFC3339Nano)
	jb.buf.WriteByte('"')
}

func (jb *jsonBuilder) appendKey(key string) {
	jb.appendString(key)
	jb.buf.WriteByte(':')
}

func (jb *jsonBuilder) appendString(v string) {
	jb.buf.WriteByte('"')
	jb.buf.WriteString(v)
	jb.buf.WriteByte('"')
}

func (jb *jsonBuilder) appendJSON(v any) {
	var bb bytes.Buffer
	enc := json.NewEncoder(&bb)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return
	}
	bs := bb.Bytes()
	jb.buf.Write(bs[:len(bs)-1]) // remove final newline

}
