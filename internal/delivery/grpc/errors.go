package grpc

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func errNotFound(kind, id string) error {
	return status.Errorf(codes.NotFound, "%s not found: %s", kind, id)
}

func errAlreadyExists(kind, id string) error {
	return status.Errorf(codes.AlreadyExists, "%s already exists: %s", kind, id)
}

func errInvalidArg(msg string) error {
	return status.Errorf(codes.InvalidArgument, "%s", msg)
}

func errInternal(err error) error {
	return status.Errorf(codes.Internal, "%v", err)
}
