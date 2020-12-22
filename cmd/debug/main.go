package main

import (
	"log"

	"github.com/trisacrypto/testnet/pkg/trisads"
	"github.com/trisacrypto/testnet/pkg/trisads/pb"
	"github.com/trisacrypto/testnet/pkg/trisads/store"
)

func main() {
	var err error
	var conf *trisads.Settings
	if conf, err = trisads.Config(); err != nil {
		log.Fatal(err)
	}

	var db store.Store
	if db, err = store.Open(conf.DatabaseDSN); err != nil {
		log.Fatal(err)
	}

	var careqs []*pb.CertificateRequest
	if careqs, err = db.ListCertRequests(); err != nil {
		log.Fatal(err)
	}

	for _, req := range careqs {
		req.Status = pb.CertificateRequestState_PROCESSING
		if err = db.SaveCertRequest(req); err != nil {
			log.Fatal(err)
		}
	}

}
