package kaobei

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

var (
	browser *rod.Browser
	once    sync.Once
)

func getBrowser(headless bool, remoteDebugAddr string) *rod.Browser {
	once.Do(func() {
		var u string

		l := launcher.New().
			Headless(headless).
			NoSandbox(true).
			Devtools(false).
			Leakless(false) //禁用 leakless

		// 如果系统有 Chrome，用系统 Chrome
		if path, found := launcher.LookPath(); found {
			l.Bin(path)
		} else {
			fmt.Println("未找到系统 Chrome，rod 将尝试下载...")
		}

		if remoteDebugAddr != "" {
			// 连接已有浏览器（端口号如 "9222"）
			l.RemoteDebuggingPort(9222) // 或 strconv.Atoi(remoteDebugAddr)
		}

		u = l.MustLaunch()

		browser = rod.New().
			ControlURL(u).
			Trace(false).
			SlowMotion(300 * time.Millisecond).
			MustConnect()
	})

	return browser
}

// InitializeBrowser 返回一个可用的 Page（类似原来的 driver）
func InitializeBrowser(headless bool, remoteDebugAddr string) *rod.Page {
	b := getBrowser(headless, remoteDebugAddr)
	page := b.MustPage("")
	page.MustSetUserAgent(&proto.NetworkSetUserAgentOverride{
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	})
	return page
}

func Login(page *rod.Page, username, password string) {
	page.MustNavigate("https://www.mangacopy.com/web/login/loginByAccount?url=person%2Fhome")
	page.MustWaitLoad().MustWaitIdle()

	// 等待表单加载
	page.Timeout(30 * time.Second).MustElement(".el-input__inner")

	// 用户名输入框
	userInput := page.MustElement(`input[type="text"].el-input__inner`)
	userInput.MustInput(username)

	// 密码输入框
	passInput := page.MustElement(`input[type="password"].el-input__inner`)
	passInput.MustInput(password)

	// 登录按钮 - 使用 MustElementR
	loginBtn, err := page.ElementR(`button`, `登錄`)
	if err != nil {
		// 尝试其他选择器
		loginBtn, err = page.Element(`.el-button.el-button--primary`)
		if err != nil {
			log.Printf("找不到登录按钮: %v", err)
			page.MustScreenshot("login_error.png")
			return
		}
	}

	loginBtn.MustClick()
	page.MustWaitNavigation()
	time.Sleep(3 * time.Second)
}

func GetComicInfo(page *rod.Page, url string) (string, int, []string) {
	fmt.Println("访问拷贝页面中...")
	page.MustNavigate(url).MustWaitLoad().MustWaitIdle()
	page.Timeout(30 * time.Second).MustElement("#default全部")
	html := page.MustHTML()

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		panic(err)
	}

	title := doc.Find("h6").First().Text()
	if title == "" {
		title = "未知漫画"
	}
	os.MkdirAll(title, 0755)
	div := doc.Find(`div[id="default全部"]`).First()
	fmt.Print(div.Length())

	var srcList []string
	div.Find("a").Each(func(_ int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		full := "https://www.mangacopy.com" + href
		srcList = append(srcList, full)
	})

	return title, len(srcList), srcList
}

func UrlProducer(b *rod.Browser, src string, count int, title string, pageNum int, ch chan map[int][]string) {
	page := b.MustPage("")
	defer page.Close()

	page.MustNavigate(src)
	page.MustWaitLoad().MustWaitIdle()
	page.Timeout(30*time.Second).WaitElementsMoreThan("body", 0)

	for i := 0; i < count; i++ {
		page.Keyboard.Press(input.Space)
		time.Sleep(100 * time.Millisecond)
	}

	html := page.MustHTML()
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))

	tempTr := doc.Find("span.comicCount").First().Text()
	tempInt, _ := strconv.Atoi(tempTr)

	temp := GetUrl(html, pageNum+1)

	fmt.Println()
	fmt.Printf("%d %d\n", tempInt, len(temp[pageNum+1]))

	if tempInt == len(temp[pageNum+1]) {
		ch <- temp
	} else {
		UrlProducer(b, src, count*2, title, pageNum, ch)
	}
}

func StartThreads(b *rod.Browser, srcList []string, startChapter, endChapter int, title string, mode int) {
	ch := make(chan map[int][]string, endChapter-startChapter+1)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for item := range ch {
			for chapter, srcs := range item {
				teNum := len(srcs)
				if mode == 1 {
					GetDownurlSep(chapter, teNum, title, srcs, endChapter)
				} else {
					GetDownurl(chapter, teNum, title, srcs, endChapter)
				}
			}
		}
	}()

	count := 4
	for i := startChapter - 1; i < endChapter && i < len(srcList); i++ {
		func(idx int) {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("章节 %d panic: %v (src: %s)\n", idx+1, r, srcList[idx])
					ch <- map[int][]string{idx + 1: []string{}}
				}
			}()
			UrlProducer(b, srcList[idx], count, title, idx, ch)
		}(i)
	}
	close(ch)
	wg.Wait()
}
