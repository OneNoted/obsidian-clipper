//go:build js && wasm

package main

import (
	"syscall/js"

	"obsidianclipper/go/filters"
)

var callbacks []js.Func

func main() {
	applyFilter := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) < 2 {
			return js.Null()
		}
		name := args[0].String()
		input := args[1].String()
		param := ""
		if len(args) > 2 && !args[2].IsUndefined() && !args[2].IsNull() {
			param = args[2].String()
		}
		result, ok := filters.Apply(name, input, param)
		if !ok {
			return js.Null()
		}
		return result
	})
	callbacks = append(callbacks, applyFilter)
	js.Global().Set("obsidianClipperGo", map[string]any{
		"applyFilter": applyFilter,
	})
	select {}
}
