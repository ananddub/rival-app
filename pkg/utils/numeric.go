package utils

import (
	"math/big"
	"strconv"

	"github.com/jackc/pgx/v5/pgtype"
)

// Float64ToNumeric converts float64 to pgtype.Numeric
func Float64ToNumeric(f float64) pgtype.Numeric {
	if f == 0 {
		return pgtype.Numeric{Int: big.NewInt(0), Exp: 0, NaN: false, Valid: true}
	}
	
	// Convert to string with 2 decimal places
	str := strconv.FormatFloat(f, 'f', 2, 64)
	
	var num pgtype.Numeric
	err := num.Scan(str)
	if err != nil {
		return pgtype.Numeric{Valid: false}
	}
	
	return num
}

// NumericToFloat64 converts pgtype.Numeric to float64
func NumericToFloat64(n pgtype.Numeric) float64 {
	if !n.Valid || n.NaN {
		return 0
	}
	
	// Convert to string then parse
	str := n.Int.String()
	if n.Exp < 0 {
		// Handle decimal places
		exp := int(-n.Exp)
		if len(str) <= exp {
			// Pad with zeros
			str = "0." + padLeft(str, exp, '0')
		} else {
			// Insert decimal point
			pos := len(str) - exp
			str = str[:pos] + "." + str[pos:]
		}
	}
	
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0
	}
	
	return f
}

func padLeft(str string, length int, pad rune) string {
	for len(str) < length {
		str = string(pad) + str
	}
	return str
}
