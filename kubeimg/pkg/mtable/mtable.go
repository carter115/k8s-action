package mtable

import (
	"fmt"
	"github.com/liushuochen/gotable"
	"github.com/liushuochen/gotable/table"
)

func GenTable(maplist []map[string]string, title []string) *table.Table {
	tb, err := gotable.Create(title...)
	if err != nil {
		fmt.Println("create tables error:", err)
		return nil
	}
	tb.AddRows(maplist)
	return tb
}
