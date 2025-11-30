package middleware

import (
	"context"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"rival/config"
	"rival/internal/auth/util"
)

// AuthInterceptor verifies JWT tokens
func AuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Skip auth for public endpoints
	if isPublicEndpoint(info.FullMethod) {
		return handler(ctx, req)
	}

	// Extract token from metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "Missing metadata")
	}

	authHeader := md.Get("authorization")
	if len(authHeader) == 0 {
		return nil, status.Error(codes.Unauthenticated, "Missing authorization header")
	}

	token := strings.TrimPrefix(authHeader[0], "Bearer ")
	if token == authHeader[0] {
		return nil, status.Error(codes.Unauthenticated, "Invalid authorization format")
	}

	// Verify JWT token
	cfg := config.GetConfig()
	jwtUtil := util.NewJWTUtil(cfg.JWT.Secret, time.Duration(cfg.JWT.ExpiryHour)*time.Hour, 7*24*time.Hour)
	claims, err := jwtUtil.ValidateToken(token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "Invalid token")
	}

	// Add user info to context
	ctx = context.WithValue(ctx, "user_id", claims.UserID)
	ctx = context.WithValue(ctx, "email", claims.Email)

	return handler(ctx, req)
}

// isPublicEndpoint checks if endpoint requires authentication
func isPublicEndpoint(method string) bool {
	publicEndpoints := []string{
		"/api.AuthService/Signup",
		"/api.AuthService/Login",
		"/api.AuthService/VerifyOTP",
		"/api.AuthService/ResendOTP",
		"/api.AuthService/ForgotPassword",
		"/api.AuthService/FirebaseLogin",
		"/api.AuthService/ResetPassword",
		"/rival.api.v1.AuthService/Signup",
		"/rival.api.v1.AuthService/Login",
		"/rival.api.v1.AuthService/VerifyOTP",
		"/rival.api.v1.AuthService/ResendOTP",
		"/rival.api.v1.AuthService/FirebaseLogin",
		"/rival.api.v1.AuthService/ForgotPassword",
		"/rival.api.v1.AuthService/ResetPassword",
	}

	for _, endpoint := range publicEndpoints {
		if method == endpoint {
			return true
		}
	}
	return false
}
