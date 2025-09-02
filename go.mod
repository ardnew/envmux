module github.com/ardnew/envmux

go 1.25.0

require (
	github.com/ardnew/mung v0.3.0
	github.com/carlmjohnson/flowmatic v0.23.4
	github.com/expr-lang/expr v1.17.6
	github.com/peterbourgon/ff/v4 v4.0.0-alpha.4
	github.com/pkg/profile v1.7.0
	golang.org/x/text v0.28.0
)

require (
	github.com/carlmjohnson/deque v0.23.1 // indirect
	github.com/felixge/fgprof v0.9.5 // indirect
	github.com/google/pprof v0.0.0-20250830080959-101d87ff5bc3 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
)

replace github.com/peterbourgon/ff/v4 => github.com/ardnew/ff/v4 v4.0.0-alpha.4.0.20250620043230-85a893511772
