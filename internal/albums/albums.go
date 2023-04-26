package albums

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const albumUrl = "https://photoslibrary.googleapis.com/v1/albums"

type AlbumList struct {
	Albums        []AlbumData
	NextPageToken string
}
type AlbumData struct {
	Id                    string
	Title                 string
	ProductUrl            string
	CoverPhotoBaseUrl     string
	CoverPhotoMediaItemId string
	IsWriteable           string
	MediaItemsCount       string
}

func callGoogleApi(url string, token string, pageToken string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header = map[string][]string{
		"Authorization": {fmt.Sprintf("Bearer %s", token)}}

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		panic(err)

	}
	defer req.Body.Close()
	return io.ReadAll(res.Body)
}

func listAllAlbums(token string) AlbumList {
	getData := func(token string, nextToken string) AlbumList {
		body, err := callGoogleApi(albumUrl, token, nextToken)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(body))

		var data AlbumList
		errorz := json.Unmarshal(body, &data)
		if errorz != nil {
			panic(errorz)
		}
		return data
	}

	data := getData(token, "")

	//todo: figure out best way to handle pagination
	if data.NextPageToken != "" {
		data2 := getData(token, data.NextPageToken)
		fmt.Println(data2)
	}
	return data
}

type albumPost struct {
	Title string `json:"title"`
}
type postData struct {
	Album albumPost `json:"album"`
}

type createAlbumnResponse struct {
	Id          string
	Title       string
	ProductUrl  string
	IsWriteable bool
}

func createNewAlbum(token string) string {
	data := &postData{Album: albumPost{Title: "test albumn"}}

	payload, _ := json.Marshal(data)

	ptoken := fmt.Sprintf("Bearer %s", token)

	req, err := http.NewRequest(http.MethodPost, albumUrl, bytes.NewBuffer(payload))
	if err != nil {
		panic(err)
	}
	req.Header.Add("Authorization", ptoken)

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var albumn createAlbumnResponse

	json.Unmarshal(body, &albumn)

	fmt.Println(albumn)
	return albumn.ProductUrl
}
