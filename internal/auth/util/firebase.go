package util

import (
	"context"
	"fmt"
	
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

type FirebaseService interface {
	VerifyToken(ctx context.Context, idToken string) (*FirebaseUser, error)
}

type FirebaseUser struct {
	UID         string
	Email       string
	Name        string
	Picture     string
	PhoneNumber string
	Provider    string
}

type firebaseService struct {
	client *auth.Client
}

func NewFirebaseService(credentialsPath string) (FirebaseService, error) {
	ctx := context.Background()
	
	opt := option.WithCredentialsFile(credentialsPath)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing firebase app: %v", err)
	}
	
	client, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting firebase auth client: %v", err)
	}
	
	return &firebaseService{client: client}, nil
}

func (f *firebaseService) VerifyToken(ctx context.Context, idToken string) (*FirebaseUser, error) {
	token, err := f.client.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("error verifying firebase token: %v", err)
	}
	
	// Get user record for additional info
	userRecord, err := f.client.GetUser(ctx, token.UID)
	if err != nil {
		return nil, fmt.Errorf("error getting user record: %v", err)
	}
	
	// Determine provider
	provider := "email"
	if len(userRecord.ProviderUserInfo) > 0 {
		provider = userRecord.ProviderUserInfo[0].ProviderID
	}
	
	return &FirebaseUser{
		UID:         token.UID,
		Email:       userRecord.Email,
		Name:        userRecord.DisplayName,
		Picture:     userRecord.PhotoURL,
		PhoneNumber: userRecord.PhoneNumber,
		Provider:    provider,
	}, nil
}
