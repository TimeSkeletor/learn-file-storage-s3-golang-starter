package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return

	}
	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't find video", err)
		return
	}

	if userID != video.UserID {
		respondWithError(w, http.StatusUnauthorized, "Not authorized to update this video", err)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<30)
	file, header, err := r.FormFile("video")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}

	mediaType, _, err := mime.ParseMediaType(header.Header.Get("Content-Type"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Content-Type", err)
		return
	}
	if mediaType != "video/mp4" {
		respondWithError(w, http.StatusBadRequest, "Invalid file type", nil)
		return
	}

	tempFile, err := os.CreateTemp("", "tubely-upload.mp4")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating tempfile", err)
		return
	}

	defer os.Remove(tempFile.Name())
	defer file.Close()

	_, err = io.Copy(tempFile, file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error copying file", err)
		return
	}
	tempFile.Seek(0, io.SeekStart)
	defer tempFile.Close()

	newVideoID := make([]byte, 32)
	_, err = rand.Read(newVideoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to generate file name", err)
		return
	}
	newUuidString := base64.RawURLEncoding.EncodeToString(newVideoID)
	assetPath := getAssetPath(newUuidString, mediaType)

	objectParams := &s3.PutObjectInput{
		Bucket:      &cfg.s3Bucket,
		Key:         &assetPath,
		Body:        tempFile,
		ContentType: &mediaType,
	}

	_, err = cfg.s3Client.PutObject(context.Background(), objectParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error uploading to S3", err)
		return
	}

	newUrl := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", cfg.s3Bucket, cfg.s3Region, assetPath)
	video.VideoURL = &newUrl

	err = cfg.db.UpdateVideo(video)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update video", err)
		return
	}

}
