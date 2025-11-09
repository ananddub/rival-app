package userHandler

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
)

type GetPhotoRequest struct {
	FileName string `json:"file_name"`
}

type GetPhotoResponse struct {
	PhotoData string `json:"photo_data"` // Base64 encoded
	MimeType  string `json:"mime_type"`
}

//encore:api public method=GET path=/user/photo/:fileName
func GetPhoto(ctx context.Context, fileName string) (*GetPhotoResponse, error) {
	// Security: Only allow files from profile_photos directory
	filePath := filepath.Join("./uploads/profile_photos", filepath.Base(fileName))
	
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("photo not found")
	}

	// Read file
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read photo")
	}

	// Encode to base64
	encodedData := base64.StdEncoding.EncodeToString(fileData)

	// Determine MIME type based on extension
	ext := filepath.Ext(fileName)
	var mimeType string
	switch ext {
	case ".jpg", ".jpeg":
		mimeType = "image/jpeg"
	case ".png":
		mimeType = "image/png"
	case ".gif":
		mimeType = "image/gif"
	default:
		mimeType = "application/octet-stream"
	}

	return &GetPhotoResponse{
		PhotoData: encodedData,
		MimeType:  mimeType,
	}, nil
}

type ListPhotosResponse struct {
	Photos []string `json:"photos"`
}

//encore:api public method=GET path=/user/photos/:userID
func ListUserPhotos(ctx context.Context, userID int64) (*ListPhotosResponse, error) {
	uploadsDir := "./uploads/profile_photos"
	
	files, err := os.ReadDir(uploadsDir)
	if err != nil {
		return &ListPhotosResponse{Photos: []string{}}, nil
	}

	var userPhotos []string
	userIDStr := fmt.Sprintf("%d", userID)
	
	for _, file := range files {
		if !file.IsDir() {
			// Check if filename starts with userID followed by a dot (e.g., "123.jpg")
			fileName := file.Name()
			if len(fileName) > len(userIDStr)+1 && 
			   fileName[:len(userIDStr)] == userIDStr && 
			   fileName[len(userIDStr)] == '.' {
				userPhotos = append(userPhotos, fileName)
			}
		}
	}

	return &ListPhotosResponse{Photos: userPhotos}, nil
}
