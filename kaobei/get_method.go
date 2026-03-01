package kaobei

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

func GetUrl(html string, page int) map[int][]string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil
	}
	dataSrcList := []string{}
	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		if src, ok := s.Attr("data-src"); ok {
			dataSrcList = append(dataSrcList, src)
		}
	})
	return map[int][]string{page: dataSrcList}
}

func GetDownurl(page int, teNum int, title string, srcs []string, endChapter int) {
	var localWg sync.WaitGroup
	sem := make(chan struct{}, 1000)

	for j, src := range srcs {
		localWg.Add(1)
		sem <- struct{}{}

		go func(idx int, url string) {
			defer localWg.Done()
			defer func() { <-sem }()

			Download(url, teNum, idx+1, page, title)
		}(j, src)
	}
	if page == endChapter {
		localWg.Wait()
	}
}

func GetDownurlSep(page int, teNum int, title string, srcs []string, endChapter int) {
	subtitle := filepath.Join(title, fmt.Sprintf("%d", page))
	os.MkdirAll(subtitle, os.ModePerm)

	var localWg sync.WaitGroup
	sem := make(chan struct{}, 1000)

	for j, src := range srcs {
		localWg.Add(1)
		sem <- struct{}{}

		go func(idx int, url string) {
			defer localWg.Done()
			defer func() { <-sem }()

			Download(url, teNum, idx+1, page, subtitle)
		}(j, src)
	}
	if page == endChapter {
		localWg.Wait()
	}
}

func GetPage(srcs *goquery.Selection) (int, []string) {
	temp := 0
	tempList := []string{}
	srcs.Each(func(i int, s *goquery.Selection) {
		href, ok := s.Attr("href")
		if ok && strings.Contains(href, "comic") {
			tempList = append(tempList, "https://www.mangacopy.com"+href)
			temp++
		}
	})
	return temp, tempList
}

func GetValid(startChapter, endChapter, maxChapter int) bool {
	return 1 <= startChapter && startChapter <= endChapter && endChapter <= maxChapter
}
