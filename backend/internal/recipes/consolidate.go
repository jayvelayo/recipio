package recipes

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// Regex matches a leading number: mixed ("1 1/2"), fraction ("1/2"), decimal ("1.5"), integer ("2").
var numRe = regexp.MustCompile(`^(\d+\s+\d+/\d+|\d+/\d+|\d+(?:\.\d+)?)`)

// Volume units expressed in teaspoons.
var volumeToTsp = map[string]float64{
	"tsp": 1, "teaspoon": 1, "teaspoons": 1,
	"tbsp": 3, "tbs": 3, "tablespoon": 3, "tablespoons": 3,
	"fl oz": 6, "fluid oz": 6, "fluid ounce": 6, "fluid ounces": 6,
	"cup": 48, "cups": 48,
	"pt": 96, "pint": 96, "pints": 96,
	"qt": 192, "quart": 192, "quarts": 192,
	"gal": 768, "gallon": 768, "gallons": 768,
}

// Imperial weight units expressed in ounces.
var imperialWeightToOz = map[string]float64{
	"oz": 1, "ounce": 1, "ounces": 1,
	"lb": 16, "lbs": 16, "pound": 16, "pounds": 16,
}

// Metric weight units expressed in grams.
var metricWeightToG = map[string]float64{
	"g": 1, "gram": 1, "grams": 1,
	"kg": 1000, "kilogram": 1000, "kilograms": 1000,
}

// Metric volume units expressed in milliliters.
var metricVolumeToMl = map[string]float64{
	"ml": 1, "milliliter": 1, "milliliters": 1, "millilitre": 1, "millilitres": 1,
	"cl": 10, "centiliter": 10, "centiliters": 10,
	"dl": 100, "deciliter": 100, "deciliters": 100,
	"l": 1000, "liter": 1000, "liters": 1000, "litre": 1000, "litres": 1000,
}

type parsedQty struct {
	amount float64
	// system is one of: "volume", "imperial_weight", "metric_weight",
	// "metric_volume", "count:<unit>", "unknown"
	system string
	raw    string
}

func parseQty(q string) parsedQty {
	q = strings.TrimSpace(q)
	if q == "" {
		return parsedQty{raw: q, system: "unknown"}
	}

	m := numRe.FindString(q)
	if m == "" {
		return parsedQty{raw: q, system: "unknown"}
	}

	amount := parseMixedNum(m)
	unit := strings.ToLower(strings.TrimSpace(q[len(m):]))

	if conv, ok := volumeToTsp[unit]; ok {
		return parsedQty{amount: amount * conv, system: "volume", raw: q}
	}
	if conv, ok := imperialWeightToOz[unit]; ok {
		return parsedQty{amount: amount * conv, system: "imperial_weight", raw: q}
	}
	if conv, ok := metricWeightToG[unit]; ok {
		return parsedQty{amount: amount * conv, system: "metric_weight", raw: q}
	}
	if conv, ok := metricVolumeToMl[unit]; ok {
		return parsedQty{amount: amount * conv, system: "metric_volume", raw: q}
	}
	// Counted item (e.g., "3 cans", "2 cloves") — system encodes the unit so
	// only identical units get merged.
	return parsedQty{amount: amount, system: "count:" + unit, raw: q}
}

func parseMixedNum(s string) float64 {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, " "); i >= 0 {
		whole, _ := strconv.ParseFloat(strings.TrimSpace(s[:i]), 64)
		frac := parseSimpleNum(strings.TrimSpace(s[i+1:]))
		return whole + frac
	}
	return parseSimpleNum(s)
}

func parseSimpleNum(s string) float64 {
	if i := strings.Index(s, "/"); i >= 0 {
		n, _ := strconv.ParseFloat(s[:i], 64)
		d, _ := strconv.ParseFloat(s[i+1:], 64)
		if d == 0 {
			return 0
		}
		return n / d
	}
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// formatAmount renders a float as a whole number, common fraction, or mixed number.
func formatAmount(f float64) string {
	type fracEntry struct {
		val float64
		str string
	}
	fracs := []fracEntry{
		{1.0 / 8, "1/8"}, {1.0 / 4, "1/4"}, {1.0 / 3, "1/3"},
		{3.0 / 8, "3/8"}, {1.0 / 2, "1/2"}, {5.0 / 8, "5/8"},
		{2.0 / 3, "2/3"}, {3.0 / 4, "3/4"}, {7.0 / 8, "7/8"},
	}

	whole := math.Floor(f)
	frac := f - whole

	if frac < 0.04 {
		return fmt.Sprintf("%.0f", whole)
	}
	for _, fe := range fracs {
		if math.Abs(frac-fe.val) < 0.04 {
			if whole >= 1 {
				return fmt.Sprintf("%.0f %s", whole, fe.str)
			}
			return fe.str
		}
	}
	// Fall back to two significant figures.
	return fmt.Sprintf("%.2g", f)
}

func formatVolume(tsp float64) string {
	switch {
	case tsp >= 12:
		cups := tsp / 48
		if cups > 1 {
			return formatAmount(cups) + " cups"
		}
		return formatAmount(cups) + " cup"
	case tsp >= 3:
		return formatAmount(tsp/3) + " tbsp"
	default:
		return formatAmount(tsp) + " tsp"
	}
}

func formatImperialWeight(oz float64) string {
	if oz >= 16 {
		return formatAmount(oz/16) + " lb"
	}
	return formatAmount(oz) + " oz"
}

func formatMetricWeight(g float64) string {
	if g >= 1000 {
		return formatAmount(g/1000) + " kg"
	}
	return formatAmount(g) + " g"
}

func formatMetricVolume(ml float64) string {
	if ml >= 1000 {
		return formatAmount(ml/1000) + " l"
	}
	return formatAmount(ml) + " ml"
}

func combineQuantities(qtys []parsedQty) string {
	var parseable []parsedQty
	var unknownRaws []string
	for _, q := range qtys {
		if q.system == "unknown" {
			if q.raw != "" {
				unknownRaws = append(unknownRaws, q.raw)
			}
		} else {
			parseable = append(parseable, q)
		}
	}

	if len(parseable) == 0 {
		return strings.Join(unknownRaws, " + ")
	}

	// Check that all parseable quantities share the same unit system.
	system := parseable[0].system
	allSame := true
	for _, q := range parseable[1:] {
		if q.system != system {
			allSame = false
			break
		}
	}

	if !allSame {
		// Incompatible units — concatenate raw strings.
		var all []string
		for _, q := range qtys {
			if q.raw != "" {
				all = append(all, q.raw)
			}
		}
		return strings.Join(all, " + ")
	}

	var total float64
	for _, q := range parseable {
		total += q.amount
	}

	var formatted string
	switch system {
	case "volume":
		formatted = formatVolume(total)
	case "imperial_weight":
		formatted = formatImperialWeight(total)
	case "metric_weight":
		formatted = formatMetricWeight(total)
	case "metric_volume":
		formatted = formatMetricVolume(total)
	default:
		// "count:<unit>"
		unit := strings.TrimPrefix(system, "count:")
		if unit == "" {
			formatted = formatAmount(total)
		} else {
			formatted = formatAmount(total) + " " + unit
		}
	}

	if len(unknownRaws) > 0 {
		formatted += " + " + strings.Join(unknownRaws, " + ")
	}
	return formatted
}

// ConsolidateIngredients merges ingredients with the same name (case-insensitive),
// converting and summing quantities within compatible unit systems.
func ConsolidateIngredients(ingredients []Ingredient) []Ingredient {
	type group struct {
		name       string
		quantities []parsedQty
	}
	groups := make(map[string]*group)
	var order []string

	for _, ing := range ingredients {
		key := strings.ToLower(strings.TrimSpace(ing.Name))
		if _, exists := groups[key]; !exists {
			groups[key] = &group{name: ing.Name}
			order = append(order, key)
		}
		groups[key].quantities = append(groups[key].quantities, parseQty(ing.Quantity))
	}

	result := make([]Ingredient, 0, len(order))
	for _, key := range order {
		g := groups[key]
		result = append(result, Ingredient{
			Name:     g.name,
			Quantity: combineQuantities(g.quantities),
		})
	}
	return result
}
