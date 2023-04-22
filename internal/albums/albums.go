package albums

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

func callGoogleApi(urlString string, token string, pageToken string) ([]byte, error) {
	reqURL, err := url.Parse(urlString)
	if err != nil {
		return nil, err
	}
	if pageToken != "" {
		values := reqURL.Query()

		values.Add("pageToken", pageToken)
	}

	ptoken := fmt.Sprintf("Bearer %s", token)
	res := &http.Request{
		Method: "GET",
		URL:    reqURL,
		Header: map[string][]string{
			"Authorization": {ptoken}},
	}

	req, err := http.DefaultClient.Do(res)
	if err != nil {
		panic(err)

	}
	defer req.Body.Close()
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
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
