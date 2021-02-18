package rvasp

import (
	"fmt"
	"time"

	pb "github.com/trisacrypto/testnet/pkg/rvasp/pb/v1"
)

type updater interface {
	send(update string, cat pb.MessageCategory) error
}

// create a stream updater for simulating live update messages
func newStreamUpdater(stream pb.TRISADemo_LiveUpdatesServer, req *pb.Command, client string) *streamUpdater {
	return &streamUpdater{
		stream:    stream,
		client:    client,
		requestID: req.Id,
	}
}

// streamUpdater holds the context for sending updates based on a single request.
type streamUpdater struct {
	stream    pb.TRISADemo_LiveUpdatesServer
	client    string
	requestID uint64
}

func (s *streamUpdater) send(update string, cat pb.MessageCategory) (err error) {
	msg := &pb.Message{
		Type:      pb.RPC_NORPC,
		Id:        s.requestID,
		Update:    update,
		Category:  cat,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	if err = s.stream.Send(msg); err != nil {
		return fmt.Errorf("could not send message to %q: %s", s.client, err)
	}
	return nil
}
