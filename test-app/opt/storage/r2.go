package storage

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

// ImageUploader defines the interface for uploading images to object storage.
type ImageUploader interface {
	Upload(ctx context.Context, key string, data []byte, contentType string) (string, error)
}

// R2Client implements ImageUploader for Cloudflare R2 (S3-compatible).
type R2Client struct {
	client    *s3.Client
	bucket    string
	publicURL string
}

// NewR2Client creates a new Cloudflare R2 storage client.
func NewR2Client(endpoint, accessKeyID, secretAccessKey, bucket, publicURL string) (*R2Client, error) {
	if endpoint == "" || accessKeyID == "" || secretAccessKey == "" || bucket == "" {
		return nil, fmt.Errorf("R2 config incomplete: endpoint, access_key_id, secret_access_key, and bucket are required")
	}

	client := s3.New(s3.Options{
		BaseEndpoint: aws.String(endpoint),
		Region:       "auto",
		Credentials:  aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
	})

	return &R2Client{
		client:    client,
		bucket:    bucket,
		publicURL: strings.TrimRight(publicURL, "/"),
	}, nil
}

// Upload puts an object in R2 and returns the public URL.
func (c *R2Client) Upload(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	_, err := c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("r2 upload: %w", err)
	}

	return fmt.Sprintf("%s/%s", c.publicURL, key), nil
}

// Allowed image content types
var allowedImageTypes = map[string]string{
	"image/jpeg": "jpg",
	"image/png":  "png",
	"image/webp": "webp",
}

// maxImageSize is 2MB decoded
const maxImageSize = 2 * 1024 * 1024

// ProcessImageInput handles the image_url field from a request:
// - Empty → returns ""
// - Starts with "http" → returns as-is (external URL, HTTPS only)
// - Base64 data URI → decodes, validates, uploads to R2, returns public URL
func ProcessImageInput(ctx context.Context, uploader ImageUploader, input string) (string, error) {
	if input == "" {
		return "", nil
	}

	// External URL — keep as-is
	if strings.HasPrefix(input, "http") {
		if !strings.HasPrefix(input, "https://") {
			return "", fmt.Errorf("only HTTPS URLs are allowed")
		}
		return input, nil
	}

	// Base64 data URI
	if !strings.HasPrefix(input, "data:image/") {
		return "", fmt.Errorf("invalid image input: must be a URL or base64 data URI")
	}

	if uploader == nil {
		return "", fmt.Errorf("image upload not configured (R2 credentials missing)")
	}

	// Parse data URI: data:image/jpeg;base64,/9j/4AAQ...
	parts := strings.SplitN(input, ",", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid base64 data URI format")
	}

	decoded, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("invalid base64 encoding: %w", err)
	}

	// Size check
	if len(decoded) > maxImageSize {
		return "", fmt.Errorf("image too large: %d bytes (max %d)", len(decoded), maxImageSize)
	}

	// Content type detection (don't trust the data URI mime)
	detectedType := http.DetectContentType(decoded[:min(512, len(decoded))])
	ext, ok := allowedImageTypes[detectedType]
	if !ok {
		return "", fmt.Errorf("unsupported image type: %s (allowed: jpeg, png, webp)", detectedType)
	}

	// Generate unique key
	key := fmt.Sprintf("products-dir/%s.%s", uuid.New().String(), ext)

	return uploader.Upload(ctx, key, decoded, detectedType)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
