package bpool_test

import (
	"fmt"
	"testing"

	"github.com/odit-bit/messenger/rabbit/rlog/internal/bpool"
	"github.com/stretchr/testify/assert"
)

func Test_(t *testing.T) {
	buf := bpool.New()
	defer buf.Free()
	buf.WriteByte('{')

	buf.WriteByte('"')
	buf.WriteString("key")
	buf.WriteByte('"')

	buf.WriteByte(':')

	buf.WriteByte('"')
	buf.WriteString("value")
	buf.WriteByte('"')

	buf.WriteByte('}')

	expected := fmt.Sprintf("{%q:%q}", "key", "value")
	assert.Equal(t, expected, string(*buf))

}
