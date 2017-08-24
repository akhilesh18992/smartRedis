package table

import (
	"fmt"
	"smartRedis/color"
)

const ALIGN_RIGHT = 1

type table struct {
	data     []tableRow
	header   []string
	cols     int
	colWidth map[int]int
	padding  int
	border   string
}

type tableRow struct {
	row []Record
}

type Record struct {
	Value     string
	ColorCode int
}

func Init() *table {
	return &table{
		data:     []tableRow{},
		header:   []string{},
		cols:     0,
		colWidth: map[int]int{},
		padding:  1,
		border:   "|",
	}
}

func (t *table) SetHeader(headers []string) {
	for i, header := range headers {
		if len(header) > t.colWidth[i] {
			t.colWidth[i] = len(header)
		}
	}
	if len(headers) > t.cols {
		t.cols = len(headers)
	}
	t.header = headers
}

func (t *table) AppendRecord(row []Record) {
	var tableRow tableRow
	for i, record := range row {
		if len(record.Value) > t.colWidth[i] {
			t.colWidth[i] = len(record.Value)
		}
		tableRow.row = append(tableRow.row, record)
	}
	if len(row) > t.cols {
		t.cols = len(row)
	}
	t.data = append(t.data, tableRow)
}

func (t *table) AppendString(row []string, colorCode int) {
	var tableRow tableRow
	for i, field := range row {
		if len(field) > t.colWidth[i] {
			t.colWidth[i] = len(field)
		}
		tableRow.row = append(tableRow.row, Record{field, colorCode})
	}
	if len(row) > t.cols {
		t.cols = len(row)
	}
	t.data = append(t.data, tableRow)
}

func (t *table) Render() {
	headerOut := "| "
	divider := "+-"
	for i, header := range t.header {
		headerOut += color.Blue(tabbedOutput(header, t.colWidth[i], ALIGN_RIGHT))
		for count := 0; count < t.colWidth[i]; count++ {
			divider += "-"
		}
		if i != len(t.header)-1 {
			headerOut += " | "
			divider += "-+-"
		}
	}
	fmt.Println(divider + " +")
	fmt.Println(headerOut + " |")
	fmt.Println(divider + " +")
	for _, data := range t.data {
		out := ""
		for i, record := range data.row {
			out += t.border + t.fieldPadding() + color.Render(record.ColorCode, 0, tabbedOutput(record.Value, t.colWidth[i], ALIGN_RIGHT)) + t.fieldPadding()
		}
		out += t.border
		fmt.Println(out)
	}
	fmt.Println(divider + " +")
}

func (t *table) fieldPadding() (s string) {
	for i := 0; i < t.padding; i++ {
		s += " "
	}
	return
}

func tabbedOutput(data string, width int, align int) string {
	l := len(data)
	var out string
	for i := 1; i <= width-l; i++ {
		out += " "
	}
	if align == ALIGN_RIGHT {
		out += data
	} else {
		out = data + out
	}
	return out
}
