package scraper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseMatchReport_Basic(t *testing.T) {
	t.Parallel()

	report, err := ParseMatchReport("<html><body><div class=\"scorebox\"></div></body></html>")

	is := assert.New(t)

	is.NoError(err)
	is.Empty(report.HomeTeam)
}

func TestExtractTwoValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		text string
		want []string
	}{
		{name: "two numbers", text: "12  8", want: []string{"12", "8"}},
		{name: "with percent", text: "55%  45%", want: []string{"55", "45"}},
		{name: "no numbers", text: "Fouls", want: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			is := assert.New(t)

			is.Equal(tt.want, extractTwoValues(tt.text))
		})
	}
}

func TestIsNumeric(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		s    string
		want bool
	}{
		{name: "number", s: "123", want: true},
		{name: "empty", s: "", want: false},
		{name: "text", s: "abc", want: false},
		{name: "mixed", s: "12a", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			is := assert.New(t)

			is.Equal(tt.want, isNumeric(tt.s))
		})
	}
}
