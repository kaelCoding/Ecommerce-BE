package cloudinary

import (
    "context"
    "log"
    "os"
    "time"

    "github.com/cloudinary/cloudinary-go/v2"
    "github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

var Cld *cloudinary.Cloudinary

// Init khởi tạo Cloudinary instance
func Init() {
    var err error
    // Lấy URL từ biến môi trường
    cldURL := os.Getenv("CLOUDINARY_URL")
    if cldURL == "" {
        log.Fatal("CLOUDINARY_URL environment variable not set")
    }

    Cld, err = cloudinary.NewFromURL(cldURL)
    if err != nil {
        log.Fatalf("Failed to initialize Cloudinary: %v", err)
    }
    Cld.Config.URL.Secure = true // Luôn sử dụng HTTPS
}

// UploadToCloudinary tải file lên Cloudinary
func UploadToCloudinary(file interface{}, folder string, publicID string) (string, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
    defer cancel()

    uploadParams := uploader.UploadParams{
        PublicID: publicID,
        Folder:   folder,
    }

    uploadResult, err := Cld.Upload.Upload(ctx, file, uploadParams)
    if err != nil {
        return "", err
    }

    return uploadResult.SecureURL, nil
}