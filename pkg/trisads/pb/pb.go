package pb

//go:generate protoc -I=../../../proto --go_out=. --go_opt=module=github.com/trisacrypto/testnet/pkg/trisads/pb --go-grpc_out=. --go-grpc_opt=module=github.com/trisacrypto/testnet/pkg/trisads/pb trisads/models/v1alpha1/models.proto trisads/api/v1alpha1/api.proto trisads/models/v1alpha1/ca.proto trisads/admin/v1alpha1/admin.proto
