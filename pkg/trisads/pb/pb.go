package pb

//go:generate protoc -I=../../../proto --go_out=. --go_opt=module=github.com/trisacrypto/testnet/pkg/trisads/pb --go-grpc_out=. --go-grpc_opt=module=github.com/trisacrypto/testnet/pkg/trisads/pb trisads/models.proto trisads/api.proto trisads/ca.proto trisads/admin.proto
