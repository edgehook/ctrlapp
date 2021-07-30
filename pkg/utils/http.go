package utils

import (
	"io/ioutil"
	"net/http"
	"time"
)

func HttpGet(sourceUrl string) (string, error) {
	timeout := time.Duration(100 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	response, err := client.Get(sourceUrl)
	if err != nil {
		return "", err
	}

	if response.Body != nil {
		defer response.Body.Close()
	}

	content, err := ioutil.ReadAll(response.Body)
	if response.StatusCode != 200 {
		return "", err
	}

	return string(content), nil
}

/*
* CheckUrlAvailable:
* check url available
 */
func CheckUrlAvailable(sourceUrl string) bool {
	timeout := time.Duration(3 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	resp, err := client.Head(sourceUrl)
	if err != nil {
		return false
	}

	if resp.StatusCode != 200 {
		return false
	}

	return true
}
