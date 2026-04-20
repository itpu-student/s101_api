package services

import (
	"context"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/itpu-student/s101_api/db"
	"github.com/itpu-student/s101_api/models"
	"github.com/itpu-student/s101_api/utils"
)

const staticDir = "./static"

var allowedUsages = map[string]bool{
	models.FileUsageAvatar: true,
	models.FileUsageReview: true,
	models.FileUsagePlace:  true,
}

func UploadFile(ctx context.Context, userID, usage string, fh *multipart.FileHeader) (*models.File, error) {
	if !allowedUsages[usage] {
		return nil, NewApiErr("bad_input", "invalid usage: %s", usage)
	}

	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(fh.Filename)), ".")
	id := utils.NewUUID()
	key := id + "." + ext
	rel := "static/" + key

	if err := os.MkdirAll(staticDir, 0o755); err != nil {
		return nil, err
	}
	if err := saveMultipart(fh, filepath.Join(staticDir, key)); err != nil {
		return nil, err
	}

	f := models.File{
		ID:        id,
		Ext:       ext,
		Path:      rel,
		Usage:     usage,
		CreatedAt: time.Now().UTC(),
		CreatedBy: userID,
	}
	if _, err := db.Files().InsertOne(ctx, f); err != nil {
		_ = os.Remove(filepath.Join(staticDir, key))
		return nil, err
	}
	return &f, nil
}

func saveMultipart(fh *multipart.FileHeader, dst string) error {
	src, err := fh.Open()
	if err != nil {
		return err
	}
	defer src.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, src)
	return err
}
