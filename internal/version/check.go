package version

import (
	"encoding/json"
	"net/http"
)

type ReleaseResponse struct {
	TagName string `json:"tag_name"`
}

func FetchLatestVersion() (string, error) {
	resp, err := http.Get("https://api.github.com/repos/xdagiz/xytz/releases/latest")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var release ReleaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	return release.TagName, nil
}
