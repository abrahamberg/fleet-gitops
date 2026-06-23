package output

import (
	"fmt"
	"io"
	"sort"
	"text/tabwriter"
)

type Pair struct {
	Key   string
	Value string
}

func WriteTable(w io.Writer, headers []string, rows [][]string) error {
	writer := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	if err := writeRow(writer, headers); err != nil {
		return err
	}
	for _, row := range rows {
		if err := writeRow(writer, row); err != nil {
			return err
		}
	}
	return writer.Flush()
}

func SortedPairs(values map[string]string) []Pair {
	pairs := make([]Pair, 0, len(values))
	for key, value := range values {
		pairs = append(pairs, Pair{Key: key, Value: value})
	}
	sort.Slice(pairs, func(left, right int) bool {
		return pairs[left].Key < pairs[right].Key
	})
	return pairs
}

func writeRow(w io.Writer, columns []string) error {
	for index, column := range columns {
		if index > 0 {
			if _, err := fmt.Fprint(w, "\t"); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprint(w, column); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(w)
	return err
}
