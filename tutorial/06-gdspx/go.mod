module gdspx-demo

go 1.22.3

require godot-ext/gdspx v0.0.0 // indirect

require github.com/goplus/spx v1.0.0

require (
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/image v0.18.0 // indirect
	golang.org/x/mobile v0.0.0-20220518205345-8578da9835fd // indirect
)

replace godot-ext/gdspx => ../../../

replace github.com/goplus/spx => ../../
