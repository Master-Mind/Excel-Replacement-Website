package models

import (
	"fmt"
	"strconv"
	"strings"

	"gonum.org/v1/gonum/unit"
)

// copy-pasted from dbhandling package to avoid import cycle
const Pound = unit.Mass(0.45359237)
const Ounce = unit.Mass(0.02834952)
const IU = unit.Mass(0.00000001)     // International Unit, used for vitamins and some minerals, technically not valid because it varies by substance, but used here for simplicity
const Calorie = unit.Energy(4.184)   // 1 Calorie = 4.184 Joules
const Centimeter = unit.Length(0.01) // 1 cm = 0.01 m

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

func ParseMass(s string) (unit.Mass, error) {
	parts := strings.Split(s, " ")

	if len(parts) != 2 || len(parts) != 1 {
		return 0, fmt.Errorf("invalid mass string: %s", s)
	}

	massNum, err := strconv.ParseFloat(parts[0], 64)

	if err != nil {
		return 0, fmt.Errorf("error parsing mass: %v", err)
	}

	switch parts[1] {
	case "g":
		mass := unit.Mass(massNum) * unit.Gram
		return mass, nil
	case "kg":
		mass := unit.Mass(massNum) * unit.Kilogram
		return mass, nil
	case "mg":
		mass := unit.Mass(massNum) * unit.Gram * unit.Milli
		return mass, nil
	case "oz":
		mass := unit.Mass(massNum) * Ounce
		return mass, nil
	case "lb":
		mass := unit.Mass(massNum) * Pound
		return mass, nil
	case "µg":
		mass := unit.Mass(massNum) * unit.Gram * unit.Micro
		return mass, nil
	default:
		return 0, fmt.Errorf("unknown unit: %s", parts[1])
	}
}
