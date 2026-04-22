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
	case "calc":
		return calc(input, param), true
	case "duration":
		return duration(input, param), true
	case "callout":
		return callout(input, param), true
	case "pascal":
		return pascal(input), true
	case "remove_tags":
		return removeTags(input, param), true
	case "strip_tags":
		return stripTags(input, param), true
	case "strip_md", "stripmd":
		return stripMD(input), true
	case "nth":
		return nth(input, param), true
	case "merge":
		return merge(input, param), true
	case "footnote":
		return footnote(input), true
	case "blockquote":
		return blockquote(input), true
	case "list":
		return listFilter(input, param), true
	case "link":
		return link(input, param), true
	case "image":
		return image(input, param), true
	case "number_format":
		return numberFormat(input, param), true
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

func calc(input, param string) string {
	if param == "" {
		return input
	}
	num, err := strconv.ParseFloat(input, 64)
	if err != nil || math.IsNaN(num) {
		return input
	}
	operation := strings.TrimSpace(stripOuterQuotes(param))
	operator := ""
	valueText := ""
	if strings.HasPrefix(operation, "**") {
		operator = "**"
		valueText = operation[2:]
	} else if operation != "" {
		operator = operation[:1]
		valueText = operation[1:]
	}
	value, err := strconv.ParseFloat(valueText, 64)
	if err != nil || math.IsNaN(value) {
		return input
	}
	var result float64
	switch operator {
	case "+":
		result = num + value
	case "-":
		result = num - value
	case "*":
		result = num * value
	case "/":
		result = num / value
	case "^", "**":
		result = math.Pow(num, value)
	default:
		return input
	}
	result = math.Round(result*1e10) / 1e10
	return jsToString(result)
}

func duration(input, param string) string {
	if input == "" {
		return input
	}
	text := stripOuterQuotes(input)
	totalSeconds := 0
	iso := regexp.MustCompile(`^P(?:(\d+)Y)?(?:(\d+)M)?(?:(\d+)D)?(?:T(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?)?$`)
	matches := iso.FindStringSubmatch(text)
	if matches == nil {
		seconds, err := strconv.Atoi(text)
		if err != nil {
			return input
		}
		totalSeconds = seconds
	} else {
		multipliers := []int{365 * 24 * 3600, 30 * 24 * 3600, 24 * 3600, 3600, 60, 1}
		for i, multiplier := range multipliers {
			if matches[i+1] == "" {
				continue
			}
			value, err := strconv.Atoi(matches[i+1])
			if err != nil {
				return input
			}
			totalSeconds += value * multiplier
		}
	}
	format := param
	if format == "" {
		if totalSeconds >= 3600 {
			format = "HH:mm:ss"
		} else {
			format = "mm:ss"
		}
	}
	format = stripOuterQuotes(stripOuterParens(format))
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60
	replacements := map[string]string{
		"HH": fmt.Sprintf("%02d", hours),
		"H":  strconv.Itoa(hours),
		"mm": fmt.Sprintf("%02d", minutes),
		"m":  strconv.Itoa(minutes),
		"ss": fmt.Sprintf("%02d", seconds),
		"s":  strconv.Itoa(seconds),
	}
	return regexp.MustCompile(`HH|H|mm|m|ss|s`).ReplaceAllStringFunc(format, func(match string) string {
		return replacements[match]
	})
}

func callout(input, param string) string {
	calloutType := "info"
	title := ""
	foldState := ""
	if param != "" {
		parts := splitCommaParams(stripOuterParens(param))
		if len(parts) > 0 && parts[0] != "" {
			calloutType = parts[0]
		}
		if len(parts) > 1 && parts[1] != "" {
			title = parts[1]
		}
		if len(parts) > 2 {
			switch strings.ToLower(parts[2]) {
			case "true":
				foldState = "-"
			case "false":
				foldState = "+"
			}
		}
	}
	header := "> [!" + calloutType + "]" + foldState
	if title != "" {
		header += " " + title
	}
	lines := strings.Split(input, "\n")
	for i, line := range lines {
		lines[i] = "> " + line
	}
	return header + "\n" + strings.Join(lines, "\n")
}

func pascal(input string) string {
	var b strings.Builder
	capitalizeNext := true
	for _, r := range input {
		if unicode.IsSpace(r) || r == '_' || r == '-' {
			capitalizeNext = true
			continue
		}
		if capitalizeNext {
			b.WriteRune(unicode.ToUpper(r))
		} else {
			b.WriteRune(r)
		}
		capitalizeNext = false
	}
	return b.String()
}

func removeTags(input, param string) string {
	if param == "" {
		return input
	}
	tags := normalizeTagParamList(param)
	if len(tags) == 0 {
		return input
	}
	pattern := `</?(?:` + strings.Join(tags, "|") + `)\b[^>]*>`
	re, err := regexp.Compile(`(?i)` + pattern)
	if err != nil {
		return input
	}
	return re.ReplaceAllString(input, "")
}

func stripTags(input, param string) string {
	keepTags := normalizeTagParamList(param)
	var result string
	if len(keepTags) == 0 {
		result = regexp.MustCompile(`</?[^>]+(>|$)`).ReplaceAllString(input, "")
	} else {
		pattern := `<(?!/?(?:` + strings.Join(keepTags, "|") + `)\b)[^>]+>`
		re, err := regexp.Compile(`(?i)` + pattern)
		if err != nil {
			return input
		}
		result = re.ReplaceAllString(input, "")
	}
	entityReplacements := []struct{ old, new string }{
		{"&nbsp;", " "}, {"&amp;", "&"}, {"&lt;", "<"}, {"&gt;", ">"}, {"&quot;", `"`}, {"&#39;", "'"},
		{"&ldquo;", `"`}, {"&rdquo;", `"`}, {"&lsquo;", "'"}, {"&rsquo;", "'"}, {"&mdash;", "—"}, {"&ndash;", "–"}, {"&hellip;", "…"},
	}
	for _, replacement := range entityReplacements {
		result = strings.ReplaceAll(result, replacement.old, replacement.new)
	}
	result = regexp.MustCompile(`&#(\d+);`).ReplaceAllStringFunc(result, func(match string) string {
		parts := regexp.MustCompile(`\d+`).FindString(match)
		code, err := strconv.Atoi(parts)
		if err != nil {
			return match
		}
		return string(rune(code))
	})
	result = regexp.MustCompile(`&#x([0-9A-Fa-f]+);`).ReplaceAllStringFunc(result, func(match string) string {
		hex := regexp.MustCompile(`[0-9A-Fa-f]+`).FindString(strings.TrimPrefix(match, "&#x"))
		code, err := strconv.ParseInt(hex, 16, 32)
		if err != nil {
			return match
		}
		return string(rune(code))
	})
	result = regexp.MustCompile(`\n{3,}`).ReplaceAllString(result, "\n\n")
	return strings.TrimSpace(result)
}

func normalizeTagParamList(param string) []string {
	param = stripOuterParens(param)
	param = stripOuterQuotes(param)
	param = strings.ReplaceAll(param, `\"`, `"`)
	param = strings.ReplaceAll(param, `\'`, `'`)
	parts := strings.Split(param, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, regexp.QuoteMeta(part))
		}
	}
	return out
}

func stripMD(input string) string {
	text := input
	replacements := []struct{ pattern, replacement string }{
		{`!\[([^\]]*)\]\([^\)]+\)`, ""},
		{`!\[\[([^\]]+)\]\]`, ""},
		{`\[([^\]]+)\]\([^\)]+\)`, `$1`},
		{`https?://\S+`, ""},
		{`\*\*(.*?)\*\*`, `$1`},
		{`__(.*?)__`, `$1`},
		{`\*(.*?)\*`, `$1`},
		{`_(.*?)_`, `$1`},
		{`==(.*?)==`, `$1`},
		{`(?m)^#+\s+`, ""},
		{"`([^`]+)`", `$1`},
		{"(?s)```.*?```", ""},
		{`~~(.*?)~~`, `$1`},
		{`(?m)^[-*+] (\[[x ]\] )?`, ""},
		{`(?m)^([-*_]){3,}\s*$`, ""},
		{`(?m)^>\s+`, ""},
		{`\|.*\|`, ""},
		{`~(\w+)~`, `$1`},
		{`\^(\w+)\^`, `$1`},
		{`:[a-z_]+:`, ""},
		{`<[^>]+>`, ""},
		{`\[\s*\]`, ""},
		{`\[\^[^\]]+\]`, ""},
		{`(?m)^\*\[[^\]]+\]:.+$`, ""},
		{`\[\[([^\]|]+)\|?([^\]]*)\]\]`, `${2}${1}`},
	}
	for _, replacement := range replacements {
		re := regexp.MustCompile(replacement.pattern)
		if replacement.pattern == `\[\[([^\]|]+)\|?([^\]]*)\]\]` {
			text = re.ReplaceAllStringFunc(text, func(match string) string {
				parts := re.FindStringSubmatch(match)
				if len(parts) >= 3 && parts[2] != "" {
					return parts[2]
				}
				if len(parts) >= 2 {
					return parts[1]
				}
				return match
			})
			continue
		}
		text = re.ReplaceAllString(text, replacement.replacement)
	}
	text = regexp.MustCompile(`\n{3,}`).ReplaceAllString(text, "\n\n")
	return strings.TrimSpace(text)
}

func nth(input, param string) string {
	if input == "" || input == "undefined" || input == "null" {
		return input
	}
	var data []any
	if err := json.Unmarshal([]byte(input), &data); err != nil {
		return input
	}
	if param == "" {
		return jsonString(data)
	}
	if strings.Contains(param, ":") {
		pieces := strings.SplitN(param, ":", 2)
		positions := map[int]bool{}
		for _, raw := range strings.Split(pieces[0], ",") {
			pos, err := strconv.Atoi(strings.TrimSpace(raw))
			if err == nil && pos > 0 {
				positions[pos] = true
			}
		}
		basis, err := strconv.Atoi(strings.TrimSpace(pieces[1]))
		if err != nil || basis < 1 {
			return input
		}
		out := make([]any, 0, len(data))
		for i, item := range data {
			if positions[(i%basis)+1] {
				out = append(out, item)
			}
		}
		return jsonString(out)
	}
	expr := strings.TrimSpace(param)
	out := make([]any, 0, len(data))
	if regexp.MustCompile(`^\d+$`).MatchString(expr) {
		position, _ := strconv.Atoi(expr)
		for i, item := range data {
			if i+1 == position {
				out = append(out, item)
			}
		}
		return jsonString(out)
	}
	if regexp.MustCompile(`^\d+n$`).MatchString(expr) {
		multiplier, _ := strconv.Atoi(strings.TrimSuffix(expr, "n"))
		if multiplier == 0 {
			return input
		}
		for i, item := range data {
			if (i+1)%multiplier == 0 {
				out = append(out, item)
			}
		}
		return jsonString(out)
	}
	if match := regexp.MustCompile(`^n\+(\d+)$`).FindStringSubmatch(expr); match != nil {
		offset, _ := strconv.Atoi(match[1])
		for i, item := range data {
			if i+1 >= offset {
				out = append(out, item)
			}
		}
		return jsonString(out)
	}
	return input
}

func merge(input, param string) string {
	if input == "" || input == "undefined" || input == "null" {
		return "[]"
	}
	var array []any
	if err := json.Unmarshal([]byte(input), &array); err != nil {
		return input
	}
	if param == "" {
		return jsonString(array)
	}
	for _, item := range splitCommaParams(stripOuterParens(param)) {
		array = append(array, item)
	}
	return jsonString(array)
}

func footnote(input string) string {
	if input == "" {
		return input
	}
	var data []any
	if err := json.Unmarshal([]byte(input), &data); err == nil {
		lines := make([]string, len(data))
		for i, item := range data {
			lines[i] = fmt.Sprintf("[^%d]: %s", i+1, jsToString(item))
		}
		return strings.Join(lines, "\n\n")
	}
	return input
}

func blockquote(input string) string {
	var data []any
	if err := json.Unmarshal([]byte(input), &data); err == nil {
		return blockquoteArray(data, 1)
	}
	var scalar any
	if err := json.Unmarshal([]byte(input), &scalar); err == nil {
		return blockquoteString(jsToString(scalar), 1)
	}
	return blockquoteString(input, 1)
}

func blockquoteArray(data []any, depth int) string {
	lines := make([]string, 0, len(data))
	for _, item := range data {
		if nested, ok := item.([]any); ok {
			lines = append(lines, blockquoteArray(nested, depth+1))
		} else {
			lines = append(lines, blockquoteString(jsToString(item), depth))
		}
	}
	return strings.Join(lines, "\n")
}

func blockquoteString(input string, depth int) string {
	prefix := strings.Repeat("> ", depth)
	lines := strings.Split(input, "\n")
	for i, line := range lines {
		lines[i] = prefix + line
	}
	return strings.Join(lines, "\n")
}

func listFilter(input, param string) string {
	if input == "" {
		return input
	}
	var parsed any
	if err := json.Unmarshal([]byte(input), &parsed); err != nil {
		return listItem(input, param, 0)
	}
	if array, ok := parsed.([]any); ok {
		return listArray(array, param, 0)
	}
	return listArray([]any{parsed}, param, 0)
}

func listArray(array []any, listType string, depth int) string {
	lines := make([]string, len(array))
	for i, item := range array {
		line := listItem(item, listType, depth)
		if listType == "numbered" || listType == "numbered-task" {
			line = regexp.MustCompile(`^\t*\d+`).ReplaceAllStringFunc(line, func(match string) string {
				return strings.Repeat("\t", strings.Count(match, "\t")) + strconv.Itoa(i+1)
			})
		}
		lines[i] = line
	}
	return strings.Join(lines, "\n")
}

func listItem(item any, listType string, depth int) string {
	if nested, ok := item.([]any); ok {
		return listArray(nested, listType, depth+1)
	}
	prefix := "- "
	switch listType {
	case "numbered":
		prefix = "1. "
	case "task":
		prefix = "- [ ] "
	case "numbered-task":
		prefix = "1. [ ] "
	}
	return strings.Repeat("\t", depth) + prefix + jsToString(item)
}

func link(input, param string) string {
	if strings.TrimSpace(input) == "" {
		return input
	}
	linkText := "link"
	if param != "" {
		linkText = stripOuterQuotes(stripOuterParens(param))
	}
	var data []any
	if err := json.Unmarshal([]byte(input), &data); err == nil {
		items := make([]string, len(data))
		for i, item := range data {
			text := jsToString(item)
			if text == "" {
				items[i] = ""
			} else {
				items[i] = "[" + linkText + "](" + encodeLinkURL(escapeMarkdown(text)) + ")"
			}
		}
		return strings.Join(items, "\n")
	}
	return "[" + linkText + "](" + encodeLinkURL(escapeMarkdown(input)) + ")"
}

func image(input, param string) string {
	if strings.TrimSpace(input) == "" {
		return input
	}
	altText := ""
	if param != "" {
		altText = stripOuterQuotes(stripOuterParens(param))
	}
	var data []any
	if err := json.Unmarshal([]byte(input), &data); err == nil {
		items := make([]string, len(data))
		for i, item := range data {
			text := jsToString(item)
			if text == "" {
				items[i] = ""
			} else {
				items[i] = "![" + altText + "](" + escapeMarkdown(text) + ")"
			}
		}
		return jsonString(items)
	}
	return "![" + altText + "](" + escapeMarkdown(input) + ")"
}

func numberFormat(input, param string) string {
	decimals := 0
	decPoint := "."
	thousandsSep := ","
	if param != "" {
		parts := splitCommaParams(stripOuterParens(param))
		if len(parts) > 0 {
			if parsed, err := strconv.Atoi(parts[0]); err == nil {
				decimals = parsed
			}
		}
		if len(parts) > 1 {
			decPoint = unescapeParam(parts[1])
		}
		if len(parts) > 2 {
			thousandsSep = unescapeParam(parts[2])
		}
	}
	var parsed any
	if err := json.Unmarshal([]byte(input), &parsed); err != nil {
		parsed = input
	}
	result := numberFormatValue(parsed, decimals, decPoint, thousandsSep)
	if text, ok := result.(string); ok {
		return text
	}
	return jsonString(result)
}

func numberFormatValue(value any, decimals int, decPoint, thousandsSep string) any {
	switch v := value.(type) {
	case float64:
		return formatNumber(v, decimals, decPoint, thousandsSep)
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return v
		}
		return formatNumber(parsed, decimals, decPoint, thousandsSep)
	case []any:
		out := make([]any, len(v))
		for i, item := range v {
			out[i] = numberFormatValue(item, decimals, decPoint, thousandsSep)
		}
		return out
	default:
		return v
	}
}

func formatNumber(value float64, decimals int, decPoint, thousandsSep string) string {
	text := strconv.FormatFloat(value, 'f', decimals, 64)
	parts := strings.SplitN(text, ".", 2)
	integer := parts[0]
	sign := ""
	if strings.HasPrefix(integer, "-") {
		sign = "-"
		integer = strings.TrimPrefix(integer, "-")
	}
	var grouped []string
	for len(integer) > 3 {
		grouped = append([]string{integer[len(integer)-3:]}, grouped...)
		integer = integer[:len(integer)-3]
	}
	grouped = append([]string{integer}, grouped...)
	result := sign + strings.Join(grouped, thousandsSep)
	if len(parts) > 1 {
		result += decPoint + parts[1]
	}
	return result
}

func escapeMarkdown(input string) string {
	return strings.ReplaceAll(strings.ReplaceAll(input, "[", `\[`), "]", `\]`)
}

func encodeLinkURL(input string) string {
	return strings.ReplaceAll(input, " ", "%20")
}

func splitCommaParams(input string) []string {
	parts := []string{}
	var current strings.Builder
	inQuote := rune(0)
	escapeNext := false
	for _, r := range input {
		if escapeNext {
			current.WriteRune(r)
			escapeNext = false
			continue
		}
		if r == '\\' {
			current.WriteRune(r)
			escapeNext = true
			continue
		}
		if (r == '\'' || r == '"') && inQuote == 0 {
			inQuote = r
			current.WriteRune(r)
			continue
		}
		if r == inQuote {
			inQuote = 0
			current.WriteRune(r)
			continue
		}
		if r == ',' && inQuote == 0 {
			parts = append(parts, strings.TrimSpace(stripOuterQuotes(current.String())))
			current.Reset()
			continue
		}
		current.WriteRune(r)
	}
	if current.Len() > 0 {
		parts = append(parts, strings.TrimSpace(stripOuterQuotes(current.String())))
	}
	return parts
}

func unescapeParam(input string) string {
	return regexp.MustCompile(`\\(.)`).ReplaceAllString(input, `$1`)
}
