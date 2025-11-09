package userHandler

import "context"

type TestUserResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

//encore:api public method=GET path=/user/test
func TestUser(ctx context.Context) (*TestUserResponse, error) {
	return &TestUserResponse{
		Message: "User system is working properly",
		Status:  "success",
	}, nil
}

type UserDashboardResponse struct {
	User            UserResponse `json:"user"`
	PhotosCount     int          `json:"photos_count"`
	VerificationStatus struct {
		Email bool `json:"email"`
		Phone bool `json:"phone"`
	} `json:"verification_status"`
}

//encore:api public method=GET path=/user/dashboard/:userID
func GetUserDashboard(ctx context.Context, userID int64) (*UserDashboardResponse, error) {
	// Get user profile
	user, err := userService.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get user photos count
	photos, err := ListUserPhotos(ctx, userID)
	if err != nil {
		photos = &ListPhotosResponse{Photos: []string{}}
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

	response := &UserDashboardResponse{
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
		PhotosCount: len(photos.Photos),
	}

	response.VerificationStatus.Email = user.IsEmailVerified
	response.VerificationStatus.Phone = user.IsPhoneVerified

	return response, nil
}
