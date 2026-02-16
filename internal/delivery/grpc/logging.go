package grpc

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

func loggingUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	peerAddr := "unknown"
	if p, ok := peer.FromContext(ctx); ok && p.Addr != nil {
		peerAddr = p.Addr.String()
	}
	log.Printf("[gRPC] %s from %s", info.FullMethod, peerAddr)
	resp, err := handler(ctx, req)
	if err != nil {
		log.Printf("[gRPC] %s error: %v", info.FullMethod, err)
		return resp, err
	}
	log.Printf("[gRPC] %s ok", info.FullMethod)
	return resp, nil
}

// LoggingUnaryInterceptor возвращает параметр сервера, который регистрирует каждый унарный RPC-вызов.
func LoggingUnaryInterceptor() grpc.ServerOption {
	return grpc.UnaryInterceptor(loggingUnaryInterceptor)
}
