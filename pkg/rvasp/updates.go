package rvasp

import (
	"fmt"
	"sync"
	"time"

	pb "github.com/trisacrypto/testnet/pkg/rvasp/pb/v1"
)

// UpdateManager sends update messages to all connected clients.
type UpdateManager struct {
	sync.RWMutex
	streams map[string]pb.TRISADemo_LiveUpdatesServer
}

// NewUpdateManager creates a new update manager ready to work. For thread safety, this
// is the only object that can send messages on update streams.
func NewUpdateManager() *UpdateManager {
	return &UpdateManager{
		streams: make(map[string]pb.TRISADemo_LiveUpdatesServer),
	}
}

// Add a new client update stream.
func (u *UpdateManager) Add(client string, stream pb.TRISADemo_LiveUpdatesServer) (err error) {
	u.Lock()
	defer u.Unlock()
	if _, ok := u.streams[client]; ok {
		return fmt.Errorf("stream for client %q already exists", client)
	}
	u.streams[client] = stream
	return nil
}

// Del an old client update stream. No-op if client odesn't exist.
func (u *UpdateManager) Del(client string) {
	u.Lock()
	delete(u.streams, client)
	u.Unlock()
}

// Broadcast a message to all streams.
func (u *UpdateManager) Broadcast(requestID uint64, update string, cat pb.MessageCategory) (err error) {
	msg := &pb.Message{
		Type:      pb.RPC_NORPC,
		Id:        requestID,
		Update:    update,
		Category:  cat,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	inactive := make([]string, 0, 1)
	errs := make([]error, 0, 1)

	u.RLock()
	for client, stream := range u.streams {
		if err = stream.Send(msg); err != nil {
			errs = append(errs, fmt.Errorf("could not send message to %q: %s", client, err))
			inactive = append(inactive, client)
		}
	}
	u.RUnlock()

	if len(inactive) > 0 {
		u.Lock()
		for _, client := range inactive {
			delete(u.streams, client)
		}
		u.Unlock()
	}

	if len(errs) == 1 {
		return errs[0]
	}

	if len(errs) > 1 {
		return fmt.Errorf("%d stream send errors occurred", len(errs))
	}

	return nil
}

// Send a message to a specific stream
func (u *UpdateManager) Send(client string, msg *pb.Message) (err error) {
	u.RLock()
	stream, ok := u.streams[client]
	if !ok {
		return fmt.Errorf("no stream for client %q", client)
	}
	if err = stream.Send(msg); err != nil {
		// Does not matter if another routine gets the lock before we do - the delete
		// will be a no-op since the other routine will also error and delete.
		u.RUnlock()
		u.Del(client)
		return fmt.Errorf("could not send %s to %q: %s", msg.Type.String(), client, err)
	}
	u.RUnlock()
	return nil
}

// SendTransferError to client
func (u *UpdateManager) SendTransferError(client string, id uint64, err *pb.Error) error {
	return u.Send(client, &pb.Message{
		Type:      pb.RPC_TRANSFER,
		Id:        id,
		Timestamp: time.Now().Format(time.RFC3339),
		Category:  pb.MessageCategory_ERROR,
		Reply: &pb.Message_Transfer{Transfer: &pb.TransferReply{
			Error: err,
		}},
	})
}
