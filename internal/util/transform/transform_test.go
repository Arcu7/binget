package transform_test

import (
	"iter"
	"slices"
	"testing"

	"github.com/Arcu7/binget/internal/util/transform"
	"github.com/stretchr/testify/assert"
)

func TestFindBy(t *testing.T) {
	intTT := []struct {
		name    string
		seq     func() iter.Seq[int]
		pred    func(int) bool
		wantVal int
		found   bool
	}{
		{
			name: "Found first even number",
			seq: func() iter.Seq[int] {
				data := []int{1, 3, 5, 6, 7}
				seq := slices.Values(data)

				return seq
			},
			pred: func(v int) bool {
				return v%2 == 0
			},
			wantVal: 6,
			found:   true,
		},
		{
			name: "Not found first even number",
			seq: func() iter.Seq[int] {
				data := []int{1, 3, 5, 7}
				seq := slices.Values(data)

				return seq
			},
			pred: func(v int) bool {
				return v%2 == 0
			},
			wantVal: 0,
			found:   false,
		},
	}

	for _, tt := range intTT {
		testRunnerFindBy(t, tt)
	}

	stringTT := []struct {
		name    string
		seq     func() iter.Seq[string]
		pred    func(string) bool
		wantVal string
		found   bool
	}{
		{
			name: "Find first empty string",
			seq: func() iter.Seq[string] {
				data := []string{"hello", "world", "", "go"}
				seq := slices.Values(data)

				return seq
			},
			pred: func(v string) bool {
				return v == ""
			},
			wantVal: "",
			found:   true,
		},
	}

	for _, tt := range stringTT {
		testRunnerFindBy(t, tt)
	}
}

func testRunnerFindBy[T any](t *testing.T, tc struct {
	name    string
	seq     func() iter.Seq[T]
	pred    func(T) bool
	wantVal T
	found   bool
},
) {
	t.Helper()
	t.Run(tc.name, func(t *testing.T) {
		val, found := transform.FindBy(tc.seq(), tc.pred)
		assert.Equal(t, tc.wantVal, val)
		assert.Equal(t, tc.found, found)
	})
}
