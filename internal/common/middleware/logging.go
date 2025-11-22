package middleware

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LoggingInterceptor logs all gRPC requests
func LoggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	
	log.Printf("gRPC Request: %s", info.FullMethod)
	
	resp, err := handler(ctx, req)
	
	duration := time.Since(start)
	code := codes.OK
	if err != nil {
		if st, ok := status.FromError(err); ok {
			code = st.Code()
		}
	}
	
	log.Printf("gRPC Response: %s [%v] %v", info.FullMethod, code, duration)
	
	return resp, err
}
