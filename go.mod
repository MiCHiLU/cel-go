module github.com/google/cel-go

replace (
	github.com/antlr/antlr4 => github.com/michilu/antlr4 v0.0.0-20180803091604-411960b1878f
	github.com/google/cel-spec => ./vendor/github.com/google/cel-spec
	github.com/googleapis/googleapis => ./vendor/github.com/googleapis/googleapis
)

require (
	github.com/antlr/antlr4 v0.0.0-20180728001836-7d0787e29ca8
	github.com/golang/protobuf v1.1.0
	github.com/google/cel-spec v0.0.1
	github.com/googleapis/googleapis v0.0.0-20180731164444-8dfde277c731
	golang.org/x/net v0.0.0-20180801234040-f4c29de78a2a // indirect
	golang.org/x/text v0.3.0 // indirect
	google.golang.org/genproto v0.0.0-20180731170733-daca94659cb5 // indirect
	google.golang.org/grpc v1.14.0
)
