package main

import (
	"context"
	api "main/kitex_gen/api"
)

// Server_OperationsImpl implements the last service interface defined in the IDL.
type Server_OperationsImpl struct{}

// Reply implements the Server_OperationsImpl interface.
func (s *Server_OperationsImpl) Reply(ctx context.Context, req *api.ReplyArgs_) (resp *api.ReplyReply, err error) {
	// TODO: Your code here...
	return
}
