package kaobei

import (
	"context"
	"encoding/json"
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
	"github.com/segmentio/kafka-go"
)

var (
	browser      *rod.Browser
	once         sync.Once
	producer     *kafka.Writer
	producerOnce sync.Once
)

const (
	kafkaBroker   = "localhost:9092"
	topic         = "comic-chapters"
	numPartitions = 8
	// 建立 topic 指令：
	// kafka-topics --bootstrap-server localhost:9092 --create --topic comic-chapters --partitions 8 --replication-factor 1
)

type ChapterImages struct {
	Chapter     int      `json:"chapter"`
	Images      []string `json:"images"`
	TotalImages int      `json:"total_images"`
}

func getBrowser(headless bool, remoteDebugAddr string) *rod.Browser {
	once.Do(func() {
		var u string
		l := launcher.New().
			Headless(headless).
			NoSandbox(true).
			Devtools(false).
			Leakless(false)

		if path, found := launcher.LookPath(); found {
			l.Bin(path)
		} else {
			fmt.Println("未找到系统 Chrome，rod 将尝试下载...")
		}

		if remoteDebugAddr != "" {
			l.RemoteDebuggingPort(9222)
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

	page.Timeout(30 * time.Second).MustElement(".el-input__inner")

	userInput := page.MustElement(`input[type="text"].el-input__inner`)
	userInput.MustInput(username)

	passInput := page.MustElement(`input[type="password"].el-input__inner`)
	passInput.MustInput(password)

	loginBtn, err := page.ElementR(`button`, `登錄`)
	if err != nil {
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

func FetchChapterImages(b *rod.Browser, chapterURL string, initialScrolls int, chapterIndex int) []string {
	page := b.MustPage("")
	defer page.Close()

	page.MustNavigate(chapterURL)
	page.MustWaitLoad().MustWaitIdle()

	for i := 0; i < initialScrolls; i++ {
		page.Keyboard.Press(input.Space)
		time.Sleep(120 * time.Millisecond)
	}

	html := page.MustHTML()
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	countText := doc.Find("span.comicCount").First().Text()
	expected, _ := strconv.Atoi(countText)

	imgMap := GetUrl(html, chapterIndex+1)
	actual := len(imgMap[chapterIndex+1])
	fmt.Printf("%d|%d", expected, actual)

	if actual >= expected && actual > 0 {
		return imgMap[chapterIndex+1]
	}
	return FetchChapterImages(b, chapterURL, initialScrolls*2, chapterIndex)
}

func initProducer() {
	producerOnce.Do(func() {
		producer = &kafka.Writer{
			Addr:     kafka.TCP(kafkaBroker),
			Topic:    topic,
			Balancer: nil, // 不设定 Balancer，让手动的 Partition 生效
		}
	})
}

// Kafka 消费者（consumer group 模式，可同时消费所有 partition）
func startKafkaConsumer(title string, mode int) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{kafkaBroker},
		Topic:       topic,
		GroupID:     "kaobei-downloader",
		MinBytes:    10e3,
		MaxBytes:    10e6,
		MaxWait:     500 * time.Millisecond,
		StartOffset: kafka.FirstOffset,
		// Partition: 0, // 已移除！使用 GroupID 时绝对不要指定，否则只能读取 partition 0
	})
	defer reader.Close()

	fmt.Println("[Kafka Consumer] 已启动（consumer group 模式，支持多 partition），等待消息...")

	var globalWg sync.WaitGroup

	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("[Kafka] 读取消息错误: %v", err)
			time.Sleep(3 * time.Second)
			continue
		}

		_ = reader.CommitMessages(context.Background(), msg)

		var ci ChapterImages
		if err := json.Unmarshal(msg.Value, &ci); err != nil {
			continue
		}

		fmt.Printf("%+v\n", ci)

		if ci.Chapter == -1 {
			fmt.Println("[Kafka] 收到结束消息，等待所有下载完成...")
			globalWg.Wait()
			return
		}

		globalWg.Add(1)
		go func(chapterMsg ChapterImages) {
			defer globalWg.Done()
			fmt.Printf("[Kafka] 收到章节 %d (共 %d 张)\n", chapterMsg.Chapter, len(chapterMsg.Images))
			if mode == 1 {
				GetDownurlSep(chapterMsg.Chapter, chapterMsg.TotalImages, title, chapterMsg.Images)
			} else {
				GetDownurl(chapterMsg.Chapter, chapterMsg.TotalImages, title, chapterMsg.Images)
			}
		}(ci)
	}
}

func StartDownloadWithKafka(b *rod.Browser, srcList []string, startChapter, endChapter int, title string, mode int) {
	// 启动消费者（goroutine）
	done := make(chan bool)
	go func() {
		startKafkaConsumer(title, mode)
		done <- true
	}()

	// 初始化生产者（singleton）
	initProducer()

	ctx := context.Background()

	// 推送每个章节到「具体」的 partition
	for i := startChapter; i < endChapter+1 && i < len(srcList); i++ {
		images := FetchChapterImages(b, srcList[i], 4, i)

		msg := ChapterImages{
			Chapter:     i,
			Images:      images,
			TotalImages: len(images),
		}
		bytes, _ := json.Marshal(msg)

		partition := i % numPartitions
		err := producer.WriteMessages(ctx, kafka.Message{
			Partition: partition,
			Value:     bytes,
		})

		if err != nil {
			log.Printf("推送 chapter %d 到 partition %d 失敗: %v", i, partition, err)
		} else {
			fmt.Printf("已推送章节 %d 到 partition %d (%d 张)\n", i, partition, len(images))
		}
	}

	// 发送结束消息（任意 partition 即可）
	endBytes, _ := json.Marshal(ChapterImages{Chapter: -1})
	_ = producer.WriteMessages(ctx, kafka.Message{
		Partition: 0,
		Value:     endBytes,
	})

	fmt.Println("所有章节已推送至 Kafka，等待下载完成...")
	<-done

	// 清理
	producer.Close()
	fmt.Println("Kafka Producer 已关闭")
}
