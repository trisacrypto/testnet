package trisa

//go:generate protoc -I=../../proto --go_out=. --go_opt=module=github.com/trisacrypto/testnet/pkg/trisa --go-grpc_out=. --go-grpc_opt=module=github.com/trisacrypto/testnet/pkg/trisa trisa/protocol/v1alpha1/api.proto trisa/protocol/v1alpha1/errors.proto
