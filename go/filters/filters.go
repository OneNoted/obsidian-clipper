package filters

import (
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"
)

// Apply runs a browser-template filter implemented in Go. The bool return is
// false when the filter has not been ported and the TypeScript fallback should
// be used.
func Apply(name, input, param string) (string, bool) {
	switch name {
	case "trim":
		return strings.TrimSpace(input), true
	case "lower":
		return strings.ToLower(input), true
	case "upper":
		return strings.ToUpper(input), true
	case "kebab":
		return kebab(input), true
	case "snake":
		return snake(input), true
	case "camel":
		return camel(input), true
	case "uncamel":
		return uncamel(input), true
	case "capitalize":
		return capitalize(input), true
	case "title":
		return title(input), true
	case "split":
		return split(input, param), true
	case "join":
		return join(input, param), true
	case "first":
		return first(input), true
	case "last":
		return last(input), true
	case "length":
		return length(input), true
	case "reverse":
		return reverse(input), true
	case "slice":
		return slice(input, param), true
	case "round":
		return round(input, param), true
	case "safe_name":
		return safeName(input, param), true
	case "decode_uri":
		return decodeURI(input), true
	case "unescape":
		return strings.ReplaceAll(strings.ReplaceAll(input, `\"`, `"`), `\n`, "\n"), true
	case "unique":
		return unique(input), true
	case "wikilink":
		return wikilink(input, param), true
	default:
		return "", false
	}
}

func stripOuterParens(s string) string {
	if len(s) >= 2 && strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
		return s[1 : len(s)-1]
	}
	return s
}

func stripOuterQuotes(s string) string {
	if len(s) >= 2 {
		first, last := s[0], s[len(s)-1]
		if (first == '\'' || first == '"') && first == last {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func jsonString(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprint(v)
	}
	return string(b)
}

func jsToString(v any) string {
	switch value := v.(type) {
	case string:
		return value
	case float64:
		if math.Trunc(value) == value {
			return strconv.FormatInt(int64(value), 10)
		}
		return strconv.FormatFloat(value, 'f', -1, 64)
	case bool:
		if value {
			return "true"
		}
		return "false"
	case nil:
		return ""
	case []any:
		parts := make([]string, len(value))
		for i, item := range value {
			parts[i] = jsToString(item)
		}
		return strings.Join(parts, ",")
	case map[string]any:
		return "[object Object]"
	default:
		return fmt.Sprint(value)
	}
}

func kebab(s string) string {
	re := regexp.MustCompile(`([a-z])([A-Z])`)
	s = re.ReplaceAllString(s, `$1-$2`)
	re = regexp.MustCompile(`[\s_]+`)
	return strings.ToLower(re.ReplaceAllString(s, "-"))
}

func snake(s string) string {
	re := regexp.MustCompile(`([a-z])([A-Z])`)
	s = re.ReplaceAllString(s, `${1}_${2}`)
	re = regexp.MustCompile(`[\s-]+`)
	return strings.ToLower(re.ReplaceAllString(s, "_"))
}

func camel(s string) string {
	var b strings.Builder
	capitalizeNext := false
	wroteFirst := false
	for _, r := range s {
		switch {
		case unicode.IsSpace(r) || r == '-':
			capitalizeNext = true
			continue
		case r == '_':
			capitalizeNext = false
			continue
		}
		if !wroteFirst {
			b.WriteRune(unicode.ToLower(r))
			wroteFirst = true
		} else if capitalizeNext {
			b.WriteRune(unicode.ToUpper(r))
		} else {
			b.WriteRune(r)
		}
		capitalizeNext = false
	}
	return b.String()
}

func uncamel(s string) string {
	re := regexp.MustCompile(`([a-z0-9])([A-Z])`)
	s = re.ReplaceAllString(s, `${1} ${2}`)
	re = regexp.MustCompile(`([A-Z])([A-Z][a-z])`)
	return strings.ToLower(re.ReplaceAllString(s, `${1} ${2}`))
}

func capitalizeString(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(strings.ToLower(s))
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func capitalize(input string) string {
	return capitalizeString(input)
}

var lowercaseTitleWords = map[string]bool{
	"a": true, "an": true, "the": true, "and": true, "but": true, "or": true,
	"for": true, "nor": true, "on": true, "at": true, "to": true, "from": true,
	"by": true, "in": true, "of": true,
}

func title(input string) string {
	words := strings.Fields(input)
	for i, word := range words {
		lower := strings.ToLower(word)
		if i != 0 && lowercaseTitleWords[lower] {
			words[i] = lower
			continue
		}
		words[i] = capitalizeString(word)
	}
	return strings.Join(words, " ")
}

func split(input, param string) string {
	if param == "" {
		chars := make([]string, 0, len([]rune(input)))
		for _, r := range input {
			chars = append(chars, string(r))
		}
		return jsonString(chars)
	}
	param = stripOuterQuotes(stripOuterParens(param))
	if len([]rune(param)) == 1 {
		return jsonString(strings.Split(input, param))
	}
	re, err := regexp.Compile(param)
	if err != nil {
		return jsonString([]string{input})
	}
	return jsonString(re.Split(input, -1))
}

func join(input, param string) string {
	if input == "" || input == "undefined" || input == "null" {
		return ""
	}
	var array []any
	if err := json.Unmarshal([]byte(input), &array); err != nil {
		return input
	}
	sep := ","
	if param != "" {
		sep = strings.ReplaceAll(stripOuterQuotes(param), `\n`, "\n")
	}
	parts := make([]string, len(array))
	for i, item := range array {
		parts[i] = jsToString(item)
	}
	return strings.Join(parts, sep)
}

func first(input string) string {
	if input == "" {
		return input
	}
	var array []any
	if err := json.Unmarshal([]byte(input), &array); err != nil || len(array) == 0 {
		return input
	}
	return jsToString(array[0])
}

func last(input string) string {
	if input == "" {
		return input
	}
	var array []any
	if err := json.Unmarshal([]byte(input), &array); err != nil || len(array) == 0 {
		return input
	}
	return jsToString(array[len(array)-1])
}

func length(input string) string {
	var array []any
	if err := json.Unmarshal([]byte(input), &array); err == nil {
		return strconv.Itoa(len(array))
	}
	var object map[string]any
	if err := json.Unmarshal([]byte(input), &object); err == nil {
		return strconv.Itoa(len(object))
	}
	return strconv.Itoa(len(utf16.Encode([]rune(input))))
}

func reverse(input string) string {
	if input == "" || input == "undefined" || input == "null" {
		return ""
	}
	var array []any
	if err := json.Unmarshal([]byte(input), &array); err == nil {
		for i, j := 0, len(array)-1; i < j; i, j = i+1, j-1 {
			array[i], array[j] = array[j], array[i]
		}
		return jsonString(array)
	}
	runes := []rune(input)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func slice(input, param string) string {
	if param == "" || input == "" {
		return input
	}
	parts := strings.Split(param, ",")
	parseIndex := func(idx int) *int {
		if idx >= len(parts) {
			return nil
		}
		text := strings.TrimSpace(parts[idx])
		if text == "" {
			return nil
		}
		value, err := strconv.Atoi(text)
		if err != nil {
			return nil
		}
		return &value
	}
	startPtr, endPtr := parseIndex(0), parseIndex(1)
	var array []any
	if err := json.Unmarshal([]byte(input), &array); err == nil {
		start, end := normalizeSliceBounds(len(array), startPtr, endPtr)
		sliced := array[start:end]
		if len(sliced) == 1 {
			return jsToString(sliced[0])
		}
		return jsonString(sliced)
	}
	runes := []rune(input)
	start, end := normalizeSliceBounds(len(runes), startPtr, endPtr)
	return string(runes[start:end])
}

func normalizeSliceBounds(length int, startPtr, endPtr *int) (int, int) {
	start, end := 0, length
	if startPtr != nil {
		start = *startPtr
	}
	if endPtr != nil {
		end = *endPtr
	}
	if start < 0 {
		start = length + start
	}
	if end < 0 {
		end = length + end
	}
	if start < 0 {
		start = 0
	}
	if start > length {
		start = length
	}
	if end < start {
		end = start
	}
	if end > length {
		end = length
	}
	return start, end
}

func round(input, param string) string {
	places := -1
	if param != "" {
		parsed, err := strconv.Atoi(param)
		if err != nil {
			return input
		}
		places = parsed
	}
	var parsed any
	if err := json.Unmarshal([]byte(input), &parsed); err != nil {
		parsed = input
	}
	result := roundValue(parsed, places)
	if text, ok := result.(string); ok {
		return text
	}
	return jsonString(result)
}

func roundValue(value any, places int) any {
	switch v := value.(type) {
	case float64:
		return roundFloat(v, places)
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return v
		}
		return jsToString(roundFloat(f, places))
	case []any:
		out := make([]any, len(v))
		for i, item := range v {
			out[i] = roundValue(item, places)
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, item := range v {
			out[key] = roundValue(item, places)
		}
		return out
	default:
		return v
	}
}

func roundFloat(value float64, places int) float64 {
	if places < 0 {
		return math.Floor(value + 0.5)
	}
	factor := math.Pow(10, float64(places))
	return math.Floor(value*factor+0.5) / factor
}

func safeName(input, param string) string {
	os := strings.ToLower(strings.TrimSpace(param))
	if os == "" {
		os = "default"
	}
	sanitized := regexp.MustCompile(`[#|\^\[\]]`).ReplaceAllString(input, "")
	switch os {
	case "windows":
		sanitized = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1F]`).ReplaceAllString(sanitized, "")
		sanitized = regexp.MustCompile(`(?i)^(con|prn|aux|nul|com[0-9]|lpt[0-9])(\..*)?$`).ReplaceAllString(sanitized, `_$1$2`)
		sanitized = regexp.MustCompile(`[\s.]+$`).ReplaceAllString(sanitized, "")
	case "mac":
		sanitized = regexp.MustCompile(`[/:\x00-\x1F]`).ReplaceAllString(sanitized, "")
		sanitized = regexp.MustCompile(`^\.`).ReplaceAllString(sanitized, "_")
	case "linux":
		sanitized = regexp.MustCompile(`[/\x00-\x1F]`).ReplaceAllString(sanitized, "")
		sanitized = regexp.MustCompile(`^\.`).ReplaceAllString(sanitized, "_")
	default:
		sanitized = regexp.MustCompile(`[<>:"/\\|?*:\x00-\x1F]`).ReplaceAllString(sanitized, "")
		sanitized = regexp.MustCompile(`(?i)^(con|prn|aux|nul|com[0-9]|lpt[0-9])(\..*)?$`).ReplaceAllString(sanitized, `_$1$2`)
		sanitized = regexp.MustCompile(`[\s.]+$`).ReplaceAllString(sanitized, "")
		sanitized = regexp.MustCompile(`^\.`).ReplaceAllString(sanitized, "_")
	}
	sanitized = regexp.MustCompile(`^\.+`).ReplaceAllString(sanitized, "")
	if len([]rune(sanitized)) > 245 {
		sanitized = string([]rune(sanitized)[:245])
	}
	if sanitized == "" {
		return "Untitled"
	}
	return sanitized
}

func decodeURI(input string) string {
	decoded, err := url.PathUnescape(input)
	if err != nil {
		return input
	}
	return decoded
}

func unique(input string) string {
	var array []any
	if err := json.Unmarshal([]byte(input), &array); err != nil {
		return input
	}
	seen := map[string]bool{}
	out := make([]any, 0, len(array))
	for _, item := range array {
		key := jsonString(item)
		if !seen[key] {
			seen[key] = true
			out = append(out, item)
		}
	}
	return jsonString(out)
}

func wikilink(input, param string) string {
	if strings.TrimSpace(input) == "" {
		return input
	}
	alias := ""
	if param != "" {
		alias = stripOuterQuotes(stripOuterParens(param))
	}
	var array []any
	if err := json.Unmarshal([]byte(input), &array); err == nil {
		out := make([]string, 0, len(array))
		for _, item := range array {
			if item == nil {
				out = append(out, "")
				continue
			}
			text := jsToString(item)
			if text == "" {
				out = append(out, "")
			} else if alias != "" {
				out = append(out, "[["+text+"|"+alias+"]]")
			} else {
				out = append(out, "[["+text+"]]")
			}
		}
		return jsonString(out)
	}
	var object map[string]any
	if err := json.Unmarshal([]byte(input), &object); err == nil {
		keys := make([]string, 0, len(object))
		for key := range object {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		out := make([]string, 0, len(keys))
		for _, key := range keys {
			out = append(out, "[["+key+"|"+jsToString(object[key])+"]]")
		}
		return jsonString(out)
	}
	if alias != "" {
		return "[[" + input + "|" + alias + "]]"
	}
	return "[[" + input + "]]"
}
