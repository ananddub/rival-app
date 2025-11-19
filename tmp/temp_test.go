package tmp_test

import (
	"fmt"
	"testing"

	"github.com/xwb1989/sqlparser"
)

func TestSqlParser(t *testing.T) {
	query := "insert into user (id, name, email) values (1, 'John Doe', 'john.doe@example.com')"
	stmt, err := sqlparser.Parse(query)
	if err != nil {
		panic(err)
	}

	switch stmt := stmt.(type) {
	case *sqlparser.Insert:
		// Table name
		table := stmt.Table
		fmt.Println("Table:", table)

		// Columns
		fmt.Println("Columns:")
		for _, col := range stmt.Columns {
			fmt.Println(" -", col.String())
		}
		stmts := stmt.Rows.(sqlparser.Values)
		fmt.Println("Values:")
		for _, row := range stmts {
			for _, val := range row {
				fmt.Println(" -", sqlparser.String(val))
			}
		}
	}

}
