package connection

import (
	"rival/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var iminio *minio.Client

func NewMinioClient() (*minio.Client, error) {
	if iminio != nil {
		return iminio, nil
	}
	config := config.GetConfig()
	s3 := config.S3
	client, err := minio.New(s3.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s3.AccessKey, s3.SecretKey, ""),
		Secure: s3.SSLMode,
	})
	if err != nil {
		return nil, err
	}
	iminio = client
	return client, nil
}
