package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"rival/config"
	"rival/internal/common/middleware"

	adminhandler "rival/internal/admin/handler"
	authhandler "rival/internal/auth/handler"
	merchantshandler "rival/internal/merchants/handler"
	ordershandler "rival/internal/orders/handler"
	paymentshandler "rival/internal/payments/handler"
	usershandler "rival/internal/users/handler"

	authpb "rival/gen/proto/proto/api"
)

func main() {
	config := config.GetConfig()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Server.Port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			middleware.LoggingInterceptor,
			middleware.AuthInterceptor,
		),
	)

	// Register auth service
	authHandler, err := authhandler.NewAuthHandler()
	if err != nil {
		log.Fatalf("Failed to create auth handler: %v", err)
	}
	authpb.RegisterAuthServiceServer(s, authHandler)

	// Register users service
	usersHandler, err := usershandler.NewUserHandler()
	if err != nil {
		log.Fatalf("Failed to create users handler: %v", err)
	}
	authpb.RegisterUserServiceServer(s, usersHandler)

	// Register merchants service
	merchantsHandler, err := merchantshandler.NewMerchantHandler()
	if err != nil {
		log.Fatalf("Failed to create merchants handler: %v", err)
	}
	authpb.RegisterMerchantServiceServer(s, merchantsHandler)

	// Register payments service
	paymentsHandler, err := paymentshandler.NewPaymentHandler()
	if err != nil {
		log.Fatalf("Failed to create payments handler: %v", err)
	}
	authpb.RegisterPaymentServiceServer(s, paymentsHandler)

	// Register admin service
	adminHandler, err := adminhandler.NewAdminHandler()
	if err != nil {
		log.Fatalf("Failed to create admin handler: %v", err)
	}
	authpb.RegisterAdminServiceServer(s, adminHandler)

	// Register orders service
	ordersHandler, err := ordershandler.NewOrderHandler()
	if err != nil {
		log.Fatalf("Failed to create orders handler: %v", err)
	}
	authpb.RegisterOrderServiceServer(s, ordersHandler)

	// TODO: Add offers service when handler is implemented

	// Enable reflection for grpcurl/grpc clients
	reflection.Register(s)

	log.Println("gRPC server listening on :", config.Server.Port)
	log.Println("Services: Auth, Users, Merchants, Payments, Admin, Orders")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
