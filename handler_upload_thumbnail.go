package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
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


	const maxMemory = 10 << 20
	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to parse multipart form", err)
		return
	}

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")

	imageData, err := io.ReadAll(file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to read file data", err)
		return
	}

	videoMetadata, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Video does not exist", err)
		return	
	}

	if videoMetadata.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized user", err)
		return	
	}

	thumbnail := thumbnail{
		data: imageData,
		mediaType: contentType,
	}
	videoThumbnails[videoMetadata.ID] = thumbnail

	url := fmt.Sprintf("http://%s:%s/api/thumbnails/%s", cfg.platform, cfg.port, videoID.String())

	videoMetadata.ThumbnailURL = &url

	if err := cfg.db.UpdateVideo(videoMetadata); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed updating video metadata", err)
		return	
	}

	respondWithJSON(w, http.StatusOK, videoMetadata)
}
