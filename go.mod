module github.com/ardnew/envmux

go 1.24.2

require (
	github.com/alecthomas/participle/v2 v2.1.4
	github.com/carlmjohnson/flowmatic v0.23.4
	github.com/expr-lang/expr v1.17.5
	github.com/peterbourgon/ff/v4 v4.0.0-alpha.4
	github.com/pkg/profile v1.7.0
	golang.org/x/text v0.26.0
)

require github.com/ardnew/mung v0.3.0

require (
	github.com/carlmjohnson/deque v0.23.1 // indirect
	github.com/felixge/fgprof v0.9.5 // indirect
	github.com/google/pprof v0.0.0-20250630185457-6e76a2b096b5 // indirect
	golang.org/x/sync v0.15.0
)

replace github.com/peterbourgon/ff/v4 => github.com/ardnew/ff/v4 v4.0.0-alpha.4.0.20250620043230-85a893511772
