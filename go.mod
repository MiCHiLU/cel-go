module github.com/google/cel-go

replace (
	github.com/antlr/antlr4 => github.com/michilu/antlr4 v0.0.0-20180803091604-411960b1878f
	github.com/google/cel-spec => ./mod/github.com/google/cel-spec
	github.com/googleapis/googleapis => ./mod/github.com/googleapis/googleapis
)

require (
	github.com/antlr/antlr4 v0.0.0-20180728001836-7d0787e29ca8
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b // indirect
	github.com/golang/protobuf v1.1.0
	github.com/google/cel-spec v0.0.1
	github.com/googleapis/googleapis v0.0.0-20180731164444-8dfde277c731
	golang.org/x/net v0.0.0-20180801234040-f4c29de78a2a
	golang.org/x/sync v0.0.0-20180314180146-1d60e4601c6f // indirect
	golang.org/x/sys v0.0.0-20180802203216-0ffbfd41fbef // indirect
	golang.org/x/text v0.3.0 // indirect
	google.golang.org/genproto v0.0.0-20180731170733-daca94659cb5 // indirect
	google.golang.org/grpc v1.14.0
)
