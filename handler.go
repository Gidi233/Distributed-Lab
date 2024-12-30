package main

import (
	"context"
	api "main/kitex_gen/api"
)

// Node_OperationsImpl implements the last service interface defined in the IDL.
type Node_OperationsImpl struct{}

// Request implements the Node_OperationsImpl interface.
func (s *Node_OperationsImpl) Request(ctx context.Context, req *api.RequestAsgs) (resp *api.RequestReply, err error) {
	// TODO: Your code here...
	return
}

// Release implements the Node_OperationsImpl interface.
func (s *Node_OperationsImpl) Release(ctx context.Context, req *api.ReleaseAsgs) (resp *api.ReleaseReply, err error) {
	// TODO: Your code here...
	return
}

// Reply implements the Node_OperationsImpl interface.
func (s *Node_OperationsImpl) Reply(ctx context.Context, req *api.ReplyArgs_) (resp *api.ReplyReply, err error) {
	// TODO: Your code here...
	return
}

// Election implements the Node_OperationsImpl interface.
func (s *Node_OperationsImpl) Election(ctx context.Context, req *api.ElectionAsgs) (resp *api.ElectionAsgs, err error) {
	// TODO: Your code here...
	return
}
