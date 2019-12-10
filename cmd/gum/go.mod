module github.com/smacker/gum/cmd/gum

go 1.13

replace (
	github.com/smacker/gum => ../../
	github.com/smacker/gum/uast => ../../uast
)

require (
	github.com/jessevdk/go-flags v1.4.0 // indirect
	github.com/smacker/gum/uast v0.0.0-00010101000000-000000000000 // indirect
	gopkg.in/bblfsh/client-go.v2 v2.8.9 // indirect
)
