module github.com/gustavo-iniguez-goya/opensnitch/server/cli

go 1.14

replace (
	github.com/gustavo-iniguez-goya/opensnitch/server/cli/menus => ./menus
	github.com/gustavo-iniguez-goya/opensnitch/server/cli/views => ./views
)

require (
	github.com/eiannone/keyboard v0.0.0-20200508000154-caf4b762e807 // indirect
	github.com/gustavo-iniguez-goya/opensnitch/daemon v0.0.0-20200730200456-544ce11a21cb
	github.com/gustavo-iniguez-goya/opensnitch/server/cli/menus v0.0.0-00010101000000-000000000000 // indirect
	github.com/gustavo-iniguez-goya/opensnitch/server/cli/views v0.0.0-00010101000000-000000000000
	golang.org/x/net v0.0.0-20200813134508-3edf25e44fcc // indirect
	google.golang.org/grpc v1.31.0 // indirect
)
