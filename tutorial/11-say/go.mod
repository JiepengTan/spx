module gdspx-demo

go 1.22.3

toolchain go1.23.1

require github.com/goplus/spx v1.0.0

require (
	github.com/pkg/errors v0.9.1 // indirect
	godot-ext/gdspx v0.0.0 // indirect
	golang.org/x/image v0.18.0 // indirect
	golang.org/x/mobile v0.0.0-20220518205345-8578da9835fd // indirect
)

replace (
	github.com/goplus/spx => ../../
	godot-ext/gdspx => ../../../
	golang.org/x/image => golang.org/x/image v0.0.0-20210628002857-a66eb6448b8d
	golang.org/x/mobile => golang.org/x/mobile v0.0.0-20210902104108-5d9a33257ab5
	golang.org/x/mod => golang.org/x/mod v0.5.1
	golang.org/x/tools => golang.org/x/tools v0.1.8
)