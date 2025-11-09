package userHandler

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"encore.app/internal/user/repo"
	"encore.app/internal/user/service"
)

var (
	userRepo    = repo.New()
	userService = service.New(userRepo)
)

type UserResponse struct {
	ID              int64  `json:"id"`
	FullName        string `json:"full_name"`
	Email           string `json:"email"`
	PhoneNumber     string `json:"phone_number,omitempty"`
	ProfilePhoto    string `json:"profile_photo,omitempty"`
	DOB             string `json:"dob,omitempty"`
	IsPhoneVerified bool   `json:"is_phone_verified"`
	IsEmailVerified bool   `json:"is_email_verified"`
	Role            string `json:"role"`
	CreatedAt       string `json:"created_at"`
}

//encore:api public method=GET path=/user/profile/:userID
func GetProfile(ctx context.Context, userID int64) (*UserResponse, error) {
	user, err := userService.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	dob := ""
	if user.DOB != nil {
		dob = user.DOB.Format("2006-01-02")
	}

	phone := ""
	if user.PhoneNumber != nil {
		phone = *user.PhoneNumber
	}

	photo := ""
	if user.ProfilePhoto != nil {
		photo = *user.ProfilePhoto
	}

	return &UserResponse{
		ID:              user.ID,
		FullName:        user.FullName,
		Email:           user.Email,
		PhoneNumber:     phone,
		ProfilePhoto:    photo,
		DOB:             dob,
		IsPhoneVerified: user.IsPhoneVerified,
		IsEmailVerified: user.IsEmailVerified,
		Role:            user.Role,
		CreatedAt:       user.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

type UpdateProfileRequest struct {
	UserID      int64  `json:"user_id"`
	FullName    string `json:"full_name,omitempty"`
	PhoneNumber string `json:"phone_number,omitempty"`
	DOB         string `json:"dob,omitempty"`
}

type UpdateProfileResponse struct {
	Message string       `json:"message"`
	User    UserResponse `json:"user"`
}

//encore:api public method=PUT path=/user/profile
func UpdateProfile(ctx context.Context, req *UpdateProfileRequest) (*UpdateProfileResponse, error) {
	user, err := userService.UpdateProfile(ctx, req.UserID, req.FullName, req.PhoneNumber, req.DOB)
	if err != nil {
		return nil, err
	}

	dob := ""
	if user.DOB != nil {
		dob = user.DOB.Format("2006-01-02")
	}

	phone := ""
	if user.PhoneNumber != nil {
		phone = *user.PhoneNumber
	}

	photo := ""
	if user.ProfilePhoto != nil {
		photo = *user.ProfilePhoto
	}

	return &UpdateProfileResponse{
		Message: "Profile updated successfully",
		User: UserResponse{
			ID:              user.ID,
			FullName:        user.FullName,
			Email:           user.Email,
			PhoneNumber:     phone,
			ProfilePhoto:    photo,
			DOB:             dob,
			IsPhoneVerified: user.IsPhoneVerified,
			IsEmailVerified: user.IsEmailVerified,
			Role:            user.Role,
			CreatedAt:       user.CreatedAt.Format("2006-01-02 15:04:05"),
		},
	}, nil
}

type UploadPhotoRequest struct {
	UserID    int64  `json:"user_id"`
	PhotoData string `json:"photo_data"`
	FileName  string `json:"file_name"`
}

type UploadPhotoResponse struct {
	Message   string `json:"message"`
	PhotoPath string `json:"photo_path"`
}

//encore:api public method=POST path=/user/upload-photo
func UploadPhoto(ctx context.Context, req *UploadPhotoRequest) (*UploadPhotoResponse, error) {
	// Validate file extension
	ext := strings.ToLower(filepath.Ext(req.FileName))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" {
		return nil, fmt.Errorf("invalid file type. Only jpg, jpeg, png, gif allowed")
	}

	// Decode base64 image
	imageData, err := base64.StdEncoding.DecodeString(req.PhotoData)
	if err != nil {
		return nil, fmt.Errorf("invalid image data")
	}

	// Create uploads directory if not exists
	uploadsDir := "./uploads/profile_photos"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory")
	}

	// Generate filename with user ID
	fileName := fmt.Sprintf("%d%s", req.UserID, ext)
	filePath := filepath.Join(uploadsDir, fileName)

	// Save file
	if err := os.WriteFile(filePath, imageData, 0644); err != nil {
		return nil, fmt.Errorf("failed to save image")
	}

	// Update user profile photo path
	photoPath := fmt.Sprintf("/uploads/profile_photos/%s", fileName)
	if err := userService.UpdateProfilePhoto(ctx, req.UserID, photoPath); err != nil {
		// Delete uploaded file if database update fails
		os.Remove(filePath)
		return nil, err
	}

	return &UploadPhotoResponse{
		Message:   "Photo uploaded successfully",
		PhotoPath: photoPath,
	}, nil
}

type DeletePhotoRequest struct {
	UserID int64 `json:"user_id"`
}

type DeletePhotoResponse struct {
	Message string `json:"message"`
}

//encore:api public method=DELETE path=/user/photo/:userID
func DeletePhoto(ctx context.Context, userID int64) (*DeletePhotoResponse, error) {
	if err := userService.DeleteProfilePhoto(ctx, userID); err != nil {
		return nil, err
	}

	return &DeletePhotoResponse{
		Message: "Photo deleted successfully",
	}, nil
}

type VerifyEmailRequest struct {
	UserID int64  `json:"user_id"`
	OTP    string `json:"otp"`
}

type VerifyEmailResponse struct {
	Message string `json:"message"`
}

//encore:api public method=POST path=/user/verify-email
func VerifyEmail(ctx context.Context, req *VerifyEmailRequest) (*VerifyEmailResponse, error) {
	if err := userService.VerifyEmail(ctx, req.UserID, req.OTP); err != nil {
		return nil, err
	}

	return &VerifyEmailResponse{
		Message: "Email verified successfully",
	}, nil
}

type VerifyPhoneRequest struct {
	UserID int64  `json:"user_id"`
	OTP    string `json:"otp"`
}

type VerifyPhoneResponse struct {
	Message string `json:"message"`
}

//encore:api public method=POST path=/user/verify-phone
func VerifyPhone(ctx context.Context, req *VerifyPhoneRequest) (*VerifyPhoneResponse, error) {
	if err := userService.VerifyPhone(ctx, req.UserID, req.OTP); err != nil {
		return nil, err
	}

	return &VerifyPhoneResponse{
		Message: "Phone verified successfully",
	}, nil
}
