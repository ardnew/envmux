module github.com/ardnew/envmux

go 1.24.0

require (
	github.com/alecthomas/participle/v2 v2.1.4
	github.com/alecthomas/repr v0.4.0
	github.com/blang/semver/v4 v4.0.0
	github.com/carlmjohnson/flowmatic v0.23.4
	github.com/expr-lang/expr v1.17.5
	github.com/peterbourgon/ff/v4 v4.0.0-alpha.4
	golang.org/x/text v0.26.0
)

require github.com/carlmjohnson/deque v0.23.1 // indirect

replace github.com/peterbourgon/ff/v4 => github.com/ardnew/ff/v4 v4.0.0-alpha.4.0.20250620043230-85a893511772
