package kaobei

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var headers = map[string]string{
	"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36",
}

const maxConnections = 10

var poolSema = make(chan struct{}, maxConnections)

func Download(src string, teNum int, j int, page int, title string) {
	poolSema <- struct{}{}
	defer func() { <-poolSema }()

	split := strings.Split(src, ".")
	ext := split[len(split)-1]
	imageName := fmt.Sprintf("%d_%d.%s", page, j, ext)
	imagePath := filepath.Join(title, imageName)

	if _, err := os.Stat(imagePath); err == nil {
		fmt.Printf("%s already exists.  ", imageName)
		if teNum == j {
			fmt.Println()
			fmt.Printf("章%d success\n", page)
		}
		return
	}

	fmt.Printf("%s is downloading  ", imageName)
	if teNum == j {
		fmt.Println()
		fmt.Printf("章%d successfully downloaded\n", page)
	}

	req, err := http.NewRequest("GET", src, nil)
	if err != nil {
		return
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	for err != nil || (resp != nil && resp.StatusCode != 200) {
		resp, err = client.Do(req)
	}

	defer resp.Body.Close()

	f, err := os.Create(imagePath)
	if err != nil {
		return
	}
	defer f.Close()
	io.Copy(f, resp.Body)
}
