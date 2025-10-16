module example/list

go 1.25.2

replace github.com/tsukinose81/firchy-cri-api => ../..

require github.com/tsukinose81/firchy-cri-api v0.0.0-00010101000000-000000000000

require (
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250804133106-a7a43d27e69b // indirect
	google.golang.org/grpc v1.76.0 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	k8s.io/cri-api v0.34.1 // indirect
)
