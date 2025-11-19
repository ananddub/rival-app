package tmp_test

import (
	"fmt"

	"github.com/xwb1989/sqlparser"
)

func main() {
	query := "SELECT id, name, email FROM users WHERE age > 20"

	stmt, err := sqlparser.Parse(query)
	if err != nil {
		panic(err)
	}

	switch stmt := stmt.(type) {
	case *sqlparser.Select:
		// Table name
		table := stmt.From[0].(*sqlparser.AliasedTableExpr)
		fmt.Println("Table:", table.Expr)

		// Columns
		fmt.Println("Columns:")
		for _, col := range stmt.SelectExprs {
			if c, ok := col.(*sqlparser.AliasedExpr); ok {
				fmt.Println(" -", sqlparser.String(c.Expr))
			}
		}
	}
}
