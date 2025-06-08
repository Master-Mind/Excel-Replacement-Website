package models

import (
	"fmt"

	"gonum.org/v1/gonum/unit"
)

// FormatMass pretty-prints a gonum Mass in g, mg, or kg as appropriate.
func FormatMass(m unit.Mass) string {
	kg := float64(m / unit.Kilogram)
	if kg >= 1.0 {
		return fmt.Sprintf("%.2f kg", kg)
	}
	g := float64(m / unit.Gram)
	if g >= 1.0 {
		return fmt.Sprintf("%.1f g", g)
	}
	mg := float64(m / (unit.Gram * unit.Milli))

	if mg >= 1.0 {
		return fmt.Sprintf("%.0f mg", mg)
	}

	return fmt.Sprintf("%.2f µg", mg*1000) // Convert mg to µg for display
}
