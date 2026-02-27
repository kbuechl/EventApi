package albums

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	albumsURL     = "https://photoslibrary.googleapis.com/v1/albums"
	uploadURL     = "https://photoslibrary.googleapis.com/v1/uploads"
	mediaItemsURL = "https://photoslibrary.googleapis.com/v1/mediaItems:batchCreate"
)

// ─── Types ────────────────────────────────────────────────────────────────────

type AlbumData struct {
	Id              string `json:"id"`
	Title           string `json:"title"`
	ProductUrl      string `json:"productUrl"`
	IsWriteable     bool   `json:"isWriteable"`
	MediaItemsCount string `json:"mediaItemsCount"`
}

type albumPost struct {
	Title string `json:"title"`
}

type createAlbumRequest struct {
	Album albumPost `json:"album"`
}

type createAlbumResponse struct {
	Id          string `json:"id"`
	Title       string `json:"title"`
	ProductUrl  string `json:"productUrl"`
	IsWriteable bool   `json:"isWriteable"`
}

type newMediaItem struct {
	SimpleMediaItem simpleMediaItem `json:"simpleMediaItem"`
}

type simpleMediaItem struct {
	FileName    string `json:"fileName"`
	UploadToken string `json:"uploadToken"`
}

type batchCreateRequest struct {
	AlbumID       string         `json:"albumId"`
	NewMediaItems []newMediaItem `json:"newMediaItems"`
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func bearerHeader(token string) string {
	return "Bearer " + token
}

// ─── Public API ───────────────────────────────────────────────────────────────

// CreateAlbum creates a new Google Photos album and returns its ID.
func CreateAlbum(token, title string) (string, error) {
	payload, err := json.Marshal(createAlbumRequest{Album: albumPost{Title: title}})
	if err != nil {
		return "", fmt.Errorf("error marshaling album request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, albumsURL, bytes.NewBuffer(payload))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Authorization", bearerHeader(token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error calling Google Photos API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Google Photos API error %d: %s", resp.StatusCode, string(body))
	}

	var album createAlbumResponse
	if err := json.Unmarshal(body, &album); err != nil {
		return "", fmt.Errorf("error parsing response: %w", err)
	}
	return album.Id, nil
}

// UploadPhoto uploads raw image bytes to Google Photos staging and returns an upload token.
func UploadPhoto(token, fileName string, data []byte) (string, error) {
	req, err := http.NewRequest(http.MethodPost, uploadURL, bytes.NewBuffer(data))
	if err != nil {
		return "", fmt.Errorf("error creating upload request: %w", err)
	}
	req.Header.Set("Authorization", bearerHeader(token))
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("X-Goog-Upload-Content-Type", "image/jpeg")
	req.Header.Set("X-Goog-Upload-Protocol", "raw")
	req.Header.Set("X-Goog-Upload-File-Name", fileName)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error uploading photo: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading upload response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("upload error %d: %s", resp.StatusCode, string(body))
	}
	return string(body), nil
}

// AddPhotoToAlbum creates a media item from an upload token and adds it to an album.
// Returns the Google Photo media item ID.
func AddPhotoToAlbum(token, albumID, uploadToken, fileName string) (string, error) {
	payload, err := json.Marshal(batchCreateRequest{
		AlbumID: albumID,
		NewMediaItems: []newMediaItem{
			{SimpleMediaItem: simpleMediaItem{FileName: fileName, UploadToken: uploadToken}},
		},
	})
	if err != nil {
		return "", fmt.Errorf("error marshaling media item request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, mediaItemsURL, bytes.NewBuffer(payload))
	if err != nil {
		return "", fmt.Errorf("error creating media item request: %w", err)
	}
	req.Header.Set("Authorization", bearerHeader(token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error creating media item: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading media item response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("media item error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		NewMediaItemResults []struct {
			MediaItem struct {
				Id string `json:"id"`
			} `json:"mediaItem"`
		} `json:"newMediaItemResults"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("error parsing media item response: %w", err)
	}
	if len(result.NewMediaItemResults) == 0 {
		return "", fmt.Errorf("no media items in response")
	}
	return result.NewMediaItemResults[0].MediaItem.Id, nil
}
