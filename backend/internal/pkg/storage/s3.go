package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Storage struct {
	client     *minio.Client
	bucketName string
	endpoint   string
	publicURL  string
	useSSL     bool
}

type Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	UseSSL          bool
	PublicURL       string // Optional: public URL for serving files (e.g., Cloudflare R2 dev URL)
}

func NewS3Storage(cfg *Config) (*S3Storage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	storage := &S3Storage{
		client:     client,
		bucketName: cfg.BucketName,
		endpoint:   cfg.Endpoint,
		publicURL:  cfg.PublicURL,
		useSSL:     cfg.UseSSL,
	}

	// Ensure bucket exists
	if err := storage.ensureBucket(context.Background()); err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *S3Storage) ensureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = s.client.MakeBucket(ctx, s.bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}

		// Try to set bucket policy for public read access
		// Note: This may fail on Cloudflare R2 (doesn't support SetBucketPolicy)
		// For R2, configure public access via Cloudflare dashboard instead
		policy := fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Effect": "Allow",
					"Principal": {"AWS": "*"},
					"Action": ["s3:GetObject"],
					"Resource": ["arn:aws:s3:::%s/*"]
				}
			]
		}`, s.bucketName)

		err = s.client.SetBucketPolicy(ctx, s.bucketName, policy)
		if err != nil {
			// Ignore policy error - R2 and some S3-compatible services don't support this
			// Public access should be configured via their respective dashboards
			fmt.Printf("Warning: Could not set bucket policy (this is normal for Cloudflare R2): %v\n", err)
		}
	}

	return nil
}

func (s *S3Storage) Upload(ctx context.Context, reader io.Reader, contentType string, size int64, folder string) (string, error) {
	// Generate unique filename
	ext := getExtensionFromContentType(contentType)
	filename := fmt.Sprintf("%s/%s%s", folder, uuid.New().String(), ext)

	// Upload file
	_, err := s.client.PutObject(ctx, s.bucketName, filename, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	// Return public URL
	return s.GetPublicURL(filename), nil
}

func (s *S3Storage) Delete(ctx context.Context, fileURL string) error {
	// Extract object name from URL
	objectName, err := s.extractObjectName(fileURL)
	if err != nil {
		return err
	}

	err = s.client.RemoveObject(ctx, s.bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

func (s *S3Storage) GetPublicURL(objectName string) string {
	// If a public URL is configured (e.g., Cloudflare R2 dev URL), use it
	if s.publicURL != "" {
		return fmt.Sprintf("%s/%s", strings.TrimSuffix(s.publicURL, "/"), objectName)
	}
	// Otherwise, construct URL from endpoint
	protocol := "http"
	if s.useSSL {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s", protocol, s.endpoint, s.bucketName, objectName)
}

func (s *S3Storage) GetPresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	presignedURL, err := s.client.PresignedGetObject(ctx, s.bucketName, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return presignedURL.String(), nil
}

func (s *S3Storage) extractObjectName(fileURL string) (string, error) {
	parsed, err := url.Parse(fileURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	// Remove bucket name prefix from path
	objectName := strings.TrimPrefix(parsed.Path, "/"+s.bucketName+"/")
	return objectName, nil
}

func getExtensionFromContentType(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	default:
		return ""
	}
}

// ValidateImageContentType checks if the content type is an allowed image type
func ValidateImageContentType(contentType string) bool {
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	return allowedTypes[contentType]
}

// MaxImageSize is the maximum allowed image size (10MB)
const MaxImageSize = 10 * 1024 * 1024

// GetImageFolder returns the folder path for auction images
func GetImageFolder(auctionID uuid.UUID) string {
	return path.Join("auctions", auctionID.String())
}

// GetAvatarFolder returns the folder path for user avatars
func GetAvatarFolder(userID uuid.UUID) string {
	return path.Join("avatars", userID.String())
}
