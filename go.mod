module github.com/theforeman/ygg_worker

go 1.19

require (
	github.com/subpop/go-log v0.1.1
	github.com/google/uuid v1.6.0
	github.com/redhatinsights/yggdrasil v0.4.2
	github.com/redhatinsights/yggdrasil_v0 v0.0.0-20210811162724-41397343c25b
	google.golang.org/grpc v1.58.3
)

require (
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	golang.org/x/net v0.23.0 // indirect
	golang.org/x/sys v0.22.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230711160842-782d3b101e98 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
)

replace github.com/redhatinsights/yggdrasil_v0 v0.0.0-20210811162724-41397343c25b => ./yggdrasil
