package markdown

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestByteNewlines(t *testing.T) {
	type tc struct {
		name string
		in   []byte
		out  [][]byte
	}
	tcs := []tc{
		{
			name: "empty",
			in:   []byte{},
			out: [][]byte{
				{},
			},
		},
		{
			name: "simple",
			in:   []byte("a"),
			out: [][]byte{
				[]byte("a"),
			},
		},
		{
			name: "newline",
			in:   []byte("\n"),
			out: [][]byte{
				{},
				{},
			},
		},
		{
			name: "multiple",
			in:   []byte("a\nb"),
			out: [][]byte{
				[]byte("a"),
				[]byte("b"),
			},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			res := bytes.Split(tc.in, []byte("\n"))
			require.EqualValues(t, tc.out, res)
		})
	}
}
