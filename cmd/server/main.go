package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	deliverygrpc "queue-service/internal/delivery/grpc"
	"queue-service/internal/delivery/grpc/pb"
	"queue-service/internal/repository/memory"
	"queue-service/internal/usecase"
	"queue-service/pkg/config"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfgPath := ""
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}
	log.SetFlags(log.LstdFlags)
	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	log.Printf("config loaded: grpc_port=%d", cfg.Server.GRPCPort)

	// Repositories
	topicRepo := memory.NewTopicRepository()
	queueRepo := memory.NewQueueRepository()
	msgRepo := memory.NewMessageRepository()
	subRepo := memory.NewSubscriptionRepository()
	pendingRepo := memory.NewPendingDeliveryRepository()

	// Use cases
	topicUC := usecase.NewTopicUseCase(topicRepo, queueRepo)
	publishUC := usecase.NewPublishUseCase(topicRepo, queueRepo, msgRepo, cfg.Broker.MaxMessageSize)
	subscribeUC := usecase.NewSubscriptionUseCase(subRepo, topicRepo, queueRepo, cfg.Broker.AckTimeoutSeconds)
	consumeUC := usecase.NewConsumeUseCase(subRepo, msgRepo, pendingRepo)

	// gRPC handler and server
	handler := deliverygrpc.NewBrokerHandler(topicUC, publishUC, subscribeUC, consumeUC)
	srv := grpc.NewServer(
		deliverygrpc.LoggingUnaryInterceptor(),
	)
	pb.RegisterBrokerServer(srv, handler)
	reflection.Register(srv)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.GRPCPort))
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	go func() {
		log.Printf("broker gRPC server listening on :%d", cfg.Server.GRPCPort)
		if err := srv.Serve(lis); err != nil {
			log.Fatalf("serve: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")
	srv.GracefulStop()
}
