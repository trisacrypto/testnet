package pb

//go:generate protoc -I ../../../proto/trisads --go_out=plugins=grpc:. api.proto models.proto
