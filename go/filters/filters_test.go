package filters

import "testing"

func TestApplyStringFilters(t *testing.T) {
	cases := []struct {
		name  string
		input string
		param string
		want  string
	}{
		{"trim", "  hello  ", "", "hello"},
		{"lower", "Hello", "", "hello"},
		{"upper", "Hello", "", "HELLO"},
		{"camel", "hello_world", "", "helloworld"},
		{"kebab", "helloWorld again", "", "hello-world-again"},
		{"snake", "helloWorld again", "", "hello_world_again"},
		{"uncamel", "myHTMLParser", "", "my html parser"},
		{"safe_name", `file<>:"/\\|?*name`, "windows", "filename"},
		{"split", "a,b,c", `","`, `["a","b","c"]`},
		{"join", `["a","b","c"]`, `" | "`, "a | b | c"},
		{"first", `["a","b"]`, "", "a"},
		{"first", `[{"a":1}]`, "", "[object Object]"},
		{"join", `[{"a":1},"b"]`, ",", "[object Object],b"},
		{"last", `["a","b"]`, "", "b"},
		{"length", `["a","b"]`, "", "2"},
		{"length", `é`, "", "1"},
		{"slice", `abcdef`, "1,4", "bcd"},
		{"slice", `[{"a":1}]`, "0,1", "[object Object]"},
		{"round", `3.14159`, "2", "3.14"},
		{"round", `-3.5`, "", "-3"},
		{"decode_uri", `hello%20world`, "", "hello world"},
		{"unescape", `a\nb`, "", "a\nb"},
		{"unique", `["a","b","a"]`, "", `["a","b"]`},
		{"wikilink", `page`, "", `[[page]]`},
		{"wikilink", `["page1","page2"]`, "alias", `["[[page1|alias]]","[[page2|alias]]"]`},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := Apply(tt.name, tt.input, tt.param)
			if !ok {
				t.Fatalf("filter %s was not registered", tt.name)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestUnsupportedFilter(t *testing.T) {
	if _, ok := Apply("not_ported", "x", ""); ok {
		t.Fatal("unsupported filter reported as supported")
	}
}
