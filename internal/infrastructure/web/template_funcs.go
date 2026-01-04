// Package web provides template helper functions for OpenHost.
// These functions are available in all templates and provide common
// operations like string manipulation, number formatting, and date handling.
package web

import (
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"math"
	"net/url"
	"reflect"
	"sort"
	"strings"
	"time"
)

// templateDict creates a map from alternating key-value pairs
func templateDict(values ...any) map[string]any {
	if len(values)%2 != 0 {
		return nil
	}
	dict := make(map[string]any, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			continue
		}
		dict[key] = values[i+1]
	}
	return dict
}

// templateList creates a slice from the given values
func templateList(values ...any) []any {
	return values
}

// templateSafe marks a string as safe HTML (not escaped)
func templateSafe(s string) template.HTML {
	return template.HTML(s)
}

// templateSafeURL marks a string as a safe URL
func templateSafeURL(s string) template.URL {
	return template.URL(s)
}

// templateSafeJS marks a string as safe JavaScript
func templateSafeJS(s string) template.JS {
	return template.JS(s)
}

// templateSafeCSS marks a string as safe CSS
func templateSafeCSS(s string) template.CSS {
	return template.CSS(s)
}

// templateHTML escapes HTML entities
func templateHTML(s string) string {
	return html.EscapeString(s)
}

// templateURLEncode URL-encodes a string
func templateURLEncode(s string) string {
	return url.QueryEscape(s)
}

// templateJSONEncode encodes a value as JSON
func templateJSONEncode(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}

// Comparison functions

func templateEq(a, b any) bool {
	return reflect.DeepEqual(a, b)
}

func templateNe(a, b any) bool {
	return !reflect.DeepEqual(a, b)
}

func templateLt(a, b any) bool {
	return toFloat64(a) < toFloat64(b)
}

func templateLe(a, b any) bool {
	return toFloat64(a) <= toFloat64(b)
}

func templateGt(a, b any) bool {
	return toFloat64(a) > toFloat64(b)
}

func templateGe(a, b any) bool {
	return toFloat64(a) >= toFloat64(b)
}

// Math functions

func templateAdd(a, b any) any {
	af, bf := toFloat64(a), toFloat64(b)
	// If both inputs were integers, return integer
	if isInt(a) && isInt(b) {
		return int(af + bf)
	}
	return af + bf
}

func templateSub(a, b any) any {
	af, bf := toFloat64(a), toFloat64(b)
	if isInt(a) && isInt(b) {
		return int(af - bf)
	}
	return af - bf
}

func templateMul(a, b any) any {
	af, bf := toFloat64(a), toFloat64(b)
	if isInt(a) && isInt(b) {
		return int(af * bf)
	}
	return af * bf
}

func templateDiv(a, b any) any {
	bf := toFloat64(b)
	if bf == 0 {
		return 0
	}
	return toFloat64(a) / bf
}

func templateMod(a, b any) int {
	bi := toInt(b)
	if bi == 0 {
		return 0
	}
	return toInt(a) % bi
}

// String manipulation

// templateTruncate truncates a string to a maximum length
func templateTruncate(s string, length int, suffix ...string) string {
	if len(s) <= length {
		return s
	}
	sfx := "..."
	if len(suffix) > 0 {
		sfx = suffix[0]
	}
	return s[:length-len(sfx)] + sfx
}

// templatePluralize returns singular or plural form based on count
func templatePluralize(count int, singular, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}

// Date/Time formatting

// templateFormatDate formats a time as a date string
func templateFormatDate(t time.Time, format ...string) string {
	f := "2006-01-02"
	if len(format) > 0 {
		f = format[0]
	}
	return t.Format(f)
}

// templateFormatDateTime formats a time as a datetime string
func templateFormatDateTime(t time.Time, format ...string) string {
	f := "2006-01-02 15:04:05"
	if len(format) > 0 {
		f = format[0]
	}
	return t.Format(f)
}

// templateFormatTime formats a time as a time-only string
func templateFormatTime(t time.Time, format ...string) string {
	f := "15:04"
	if len(format) > 0 {
		f = format[0]
	}
	return t.Format(f)
}

// templateTimeAgo returns a human-readable time difference
func templateTimeAgo(t time.Time) string {
	diff := time.Since(t)

	seconds := int(diff.Seconds())
	minutes := int(diff.Minutes())
	hours := int(diff.Hours())
	days := hours / 24
	months := days / 30
	years := days / 365

	switch {
	case seconds < 60:
		return "刚刚"
	case minutes < 60:
		return fmt.Sprintf("%d分钟前", minutes)
	case hours < 24:
		return fmt.Sprintf("%d小时前", hours)
	case days < 30:
		return fmt.Sprintf("%d天前", days)
	case months < 12:
		return fmt.Sprintf("%d个月前", months)
	default:
		return fmt.Sprintf("%d年前", years)
	}
}

// Number formatting

// templateFormatNumber formats a number with thousands separators
func templateFormatNumber(n any, decimals ...int) string {
	f := toFloat64(n)
	dec := 0
	if len(decimals) > 0 {
		dec = decimals[0]
	}

	// Format with decimals
	formatted := fmt.Sprintf("%.*f", dec, f)

	// Add thousands separators
	parts := strings.Split(formatted, ".")
	intPart := parts[0]
	var result strings.Builder

	negative := false
	if len(intPart) > 0 && intPart[0] == '-' {
		negative = true
		intPart = intPart[1:]
	}

	for i, c := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			result.WriteRune(',')
		}
		result.WriteRune(c)
	}

	if negative {
		result.Reset()
		result.WriteRune('-')
		for i, c := range intPart {
			if i > 0 && (len(intPart)-i)%3 == 0 {
				result.WriteRune(',')
			}
			result.WriteRune(c)
		}
	}

	if len(parts) > 1 {
		result.WriteRune('.')
		result.WriteString(parts[1])
	}

	return result.String()
}

// templateFormatCurrency formats a number as currency
func templateFormatCurrency(n any, symbol ...string) string {
	sym := "¥"
	if len(symbol) > 0 {
		sym = symbol[0]
	}
	return sym + templateFormatNumber(n, 2)
}

// templateFormatPercent formats a number as a percentage
func templateFormatPercent(n any, decimals ...int) string {
	dec := 0
	if len(decimals) > 0 {
		dec = decimals[0]
	}
	return fmt.Sprintf("%.*f%%", dec, toFloat64(n)*100)
}

// templateFormatBytes formats bytes as human-readable size
func templateFormatBytes(bytes any) string {
	b := toFloat64(bytes)
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%.0f B", b)
	}
	div, exp := float64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	units := []string{"KB", "MB", "GB", "TB", "PB", "EB"}
	return fmt.Sprintf("%.1f %s", b/div, units[exp])
}

// Conditional helpers

// templateDefault returns the default value if the first is empty
func templateDefault(defaultVal, val any) any {
	if isEmpty(val) {
		return defaultVal
	}
	return val
}

// templateCoalesce returns the first non-empty value
func templateCoalesce(values ...any) any {
	for _, v := range values {
		if !isEmpty(v) {
			return v
		}
	}
	return nil
}

// templateTernary returns trueVal if condition is true, else falseVal
func templateTernary(condition bool, trueVal, falseVal any) any {
	if condition {
		return trueVal
	}
	return falseVal
}

// Collection helpers

// templateFirst returns the first element of a slice
func templateFirst(v any) any {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice || rv.Len() == 0 {
		return nil
	}
	return rv.Index(0).Interface()
}

// templateLast returns the last element of a slice
func templateLast(v any) any {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice || rv.Len() == 0 {
		return nil
	}
	return rv.Index(rv.Len() - 1).Interface()
}

// templateSlice returns a slice of elements
func templateSlice(v any, start, end int) any {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice {
		return nil
	}
	if start < 0 {
		start = 0
	}
	if end > rv.Len() {
		end = rv.Len()
	}
	return rv.Slice(start, end).Interface()
}

// templateLen returns the length of a collection
func templateLen(v any) int {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map, reflect.String, reflect.Chan:
		return rv.Len()
	}
	return 0
}

// templateReverse reverses a slice
func templateReverse(v any) any {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice {
		return v
	}

	length := rv.Len()
	result := reflect.MakeSlice(rv.Type(), length, length)

	for i := 0; i < length; i++ {
		result.Index(i).Set(rv.Index(length - 1 - i))
	}

	return result.Interface()
}

// templateSortBy sorts a slice of maps by a key
func templateSortBy(v any, key string) any {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice {
		return v
	}

	// Convert to sortable slice
	items := make([]map[string]any, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		item := rv.Index(i).Interface()
		if m, ok := item.(map[string]any); ok {
			items[i] = m
		}
	}

	sort.Slice(items, func(i, j int) bool {
		a := fmt.Sprint(items[i][key])
		b := fmt.Sprint(items[j][key])
		return a < b
	})

	return items
}

// templateGroupBy groups a slice of maps by a key
func templateGroupBy(v any, key string) map[string][]any {
	result := make(map[string][]any)
	rv := reflect.ValueOf(v)

	if rv.Kind() != reflect.Slice {
		return result
	}

	for i := 0; i < rv.Len(); i++ {
		item := rv.Index(i).Interface()
		if m, ok := item.(map[string]any); ok {
			groupKey := fmt.Sprint(m[key])
			result[groupKey] = append(result[groupKey], item)
		}
	}

	return result
}

// templatePluck extracts a specific key from each item
func templatePluck(v any, key string) []any {
	var result []any
	rv := reflect.ValueOf(v)

	if rv.Kind() != reflect.Slice {
		return result
	}

	for i := 0; i < rv.Len(); i++ {
		item := rv.Index(i).Interface()
		if m, ok := item.(map[string]any); ok {
			result = append(result, m[key])
		}
	}

	return result
}

// templateUnique returns unique values from a slice
func templateUnique(v any) any {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice {
		return v
	}

	seen := make(map[any]bool)
	result := reflect.MakeSlice(rv.Type(), 0, rv.Len())

	for i := 0; i < rv.Len(); i++ {
		item := rv.Index(i).Interface()
		if !seen[item] {
			seen[item] = true
			result = reflect.Append(result, rv.Index(i))
		}
	}

	return result.Interface()
}

// templateIn checks if a value is in a slice
func templateIn(needle any, haystack any) bool {
	rv := reflect.ValueOf(haystack)
	if rv.Kind() != reflect.Slice {
		return false
	}

	for i := 0; i < rv.Len(); i++ {
		if reflect.DeepEqual(needle, rv.Index(i).Interface()) {
			return true
		}
	}
	return false
}

// templateNotIn checks if a value is not in a slice
func templateNotIn(needle any, haystack any) bool {
	return !templateIn(needle, haystack)
}

// templateRange creates a range of integers
func templateRange(start, end int) []int {
	if start > end {
		start, end = end, start
	}
	result := make([]int, end-start)
	for i := range result {
		result[i] = start + i
	}
	return result
}

// URL helpers

// templateURL builds a URL with query parameters
func templateURL(base string, params ...string) string {
	if len(params) == 0 {
		return base
	}

	u, err := url.Parse(base)
	if err != nil {
		return base
	}

	q := u.Query()
	for i := 0; i < len(params)-1; i += 2 {
		q.Set(params[i], params[i+1])
	}
	u.RawQuery = q.Encode()

	return u.String()
}

// templateIsActiveURL checks if a URL matches the current path
func templateIsActiveURL(current, check string) bool {
	return current == check || strings.HasPrefix(current, check+"/")
}

// Form helpers

// templateSelected returns "selected" if values match
func templateSelected(current, check any) string {
	if reflect.DeepEqual(current, check) {
		return "selected"
	}
	return ""
}

// templateChecked returns "checked" if value is true
func templateChecked(val any) string {
	if toBool(val) {
		return "checked"
	}
	return ""
}

// templateDisabled returns "disabled" if value is true
func templateDisabled(val any) string {
	if toBool(val) {
		return "disabled"
	}
	return ""
}

// Debug helpers

// templateDump outputs a variable for debugging (JSON format)
func templateDump(v any) template.HTML {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return template.HTML(fmt.Sprintf("<pre>Error: %v</pre>", err))
	}
	return template.HTML(fmt.Sprintf("<pre>%s</pre>", html.EscapeString(string(b))))
}

// templateDebug returns a debug string representation
func templateDebug(v any) string {
	return fmt.Sprintf("%#v", v)
}

// Helper functions for type conversion

func toFloat64(v any) float64 {
	switch n := v.(type) {
	case int:
		return float64(n)
	case int8:
		return float64(n)
	case int16:
		return float64(n)
	case int32:
		return float64(n)
	case int64:
		return float64(n)
	case uint:
		return float64(n)
	case uint8:
		return float64(n)
	case uint16:
		return float64(n)
	case uint32:
		return float64(n)
	case uint64:
		return float64(n)
	case float32:
		return float64(n)
	case float64:
		return n
	case string:
		var f float64
		fmt.Sscanf(n, "%f", &f)
		return f
	default:
		return 0
	}
}

func toInt(v any) int {
	return int(math.Round(toFloat64(v)))
}

func isInt(v any) bool {
	switch v.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return true
	}
	return false
}

func toBool(v any) bool {
	switch b := v.(type) {
	case bool:
		return b
	case int:
		return b != 0
	case string:
		return b != "" && b != "0" && b != "false"
	default:
		return v != nil
	}
}

func isEmpty(v any) bool {
	if v == nil {
		return true
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.String:
		return rv.Len() == 0
	case reflect.Slice, reflect.Array, reflect.Map, reflect.Chan:
		return rv.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return rv.IsNil()
	case reflect.Bool:
		return !rv.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() == 0
	}
	return false
}
