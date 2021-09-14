module github.com/wostzone/logger

go 1.14

require (
	github.com/kr/pretty v0.1.0 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/wostzone/hubclient-go v0.0.0-00010101000000-000000000000
	golang.org/x/net v0.0.0-20210226172049-e18ecbb05110 // indirect
	golang.org/x/sys v0.0.0-20210309074719-68d13333faf2 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
)

// Until Hub is stable
replace github.com/wostzone/hubclient-go => ../hubclient-go

replace github.com/wostzone/hubserve-go => ../hubserve-go
