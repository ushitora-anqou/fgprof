package fgprof

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/google/pprof/profile"
)

// Format decides how the ouput is rendered to the user.
type Format string

const (
	// FormatFolded is used by Brendan Gregg's FlameGraph utility, see
	// https://github.com/brendangregg/FlameGraph#2-fold-stacks.
	FormatFolded Format = "folded"
	// FormatPprof is used by Google's pprof utility, see
	// https://github.com/google/pprof/blob/master/proto/README.md.
	FormatPprof Format = "pprof"
)

func writeFormatFolded(w io.Writer, s map[string]int) error {
	return writeFolded(w, s)
}

func writeFormatPprof(w io.Writer, src map[string]pprofData, hz int) error {
	return toPprof(src, hz).Write(w)
}

func writeFolded(w io.Writer, s map[string]int) error {
	for _, stack := range sortedKeys(s) {
		count := s[stack]
		if _, err := fmt.Fprintf(w, "%s %d\n", stack, count); err != nil {
			return err
		}
	}
	return nil
}

func toPprof(src map[string]pprofData, hz int) *profile.Profile {
	functionID := uint64(1)
	locationID := uint64(1)
	line := int64(1)

	p := &profile.Profile{}
	m := &profile.Mapping{ID: 1, HasFunctions: true}
	p.Mapping = []*profile.Mapping{m}
	p.SampleType = []*profile.ValueType{
		{
			Type: "samples",
			Unit: "count",
		},
		{
			Type: "time",
			Unit: "nanoseconds",
		},
	}

	for stack, data := range src {
		count := data.count
		sample := &profile.Sample{
			Value: []int64{
				int64(count),
				int64(1000 * 1000 * 1000 / hz * count),
			},
		}
		for i, fnName := range strings.Split(stack, ";") {
			fr := data.stack[i]
			function := &profile.Function{
				ID:        functionID,
				Name:      fnName,
				Filename:  fr.file,
				StartLine: fr.fStartLine,
			}
			p.Function = append(p.Function, function)

			location := &profile.Location{
				ID:      locationID,
				Mapping: m,
				Line:    []profile.Line{{Function: function, Line: fr.line}},
			}
			p.Location = append(p.Location, location)
			sample.Location = append([]*profile.Location{location}, sample.Location...)

			line++

			locationID++
			functionID++
		}
		p.Sample = append(p.Sample, sample)
	}
	return p
}

func sortedKeys(s map[string]int) []string {
	keys := make([]string, len(s))
	i := 0
	for stack := range s {
		keys[i] = stack
		i++
	}
	sort.Strings(keys)
	return keys
}
