package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"kaobei/kaobei"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print(`请输入下载模式 (回车默认分文件夹):
0 -> 全部图片放在同一文件夹
1 -> 每章一个子文件夹
`)
	modeStr, _ := reader.ReadString('\n')
	modeStr = strings.TrimSpace(modeStr)
	mode := 1
	if modeStr != "" {
		mode, _ = strconv.Atoi(modeStr)
	}

	fmt.Print(`请输入章节范围 (如 114-514，回车下载全部): `)
	rangeStr, _ := reader.ReadString('\n')
	rangeStr = strings.TrimSpace(rangeStr)
	if rangeStr == "" {
		rangeStr = "1-114514"
	}
	parts := strings.Split(rangeStr, "-")
	start, _ := strconv.Atoi(parts[0])
	end, _ := strconv.Atoi(parts[1])

	// 初始化（无头模式）
	// page := kaobei.InitializeBrowser(true, "") // headless = true, 无 debuggerAddr
	page := kaobei.InitializeBrowser(false, "") // false = 可见窗口
	defer page.Close()

	kaobei.Login(page, "test_345", "123456789") // ← 改成真实账号密码

	// 漫画页面 URL（请自行修改）
	comicURL := "https://www.mangacopy.com/comic/sydsz"
	title, totalChapters, chapterURLs := kaobei.GetComicInfo(page, comicURL)

	if end > totalChapters || end == 114514 {
		end = totalChapters
	}

	browser := page.Browser()
	defer browser.MustClose()
	fmt.Printf("漫画：《%s》 共 %d 章，将下载 %d-%d 章\n", title, totalChapters, start, end)

	kaobei.StartThreads(browser, chapterURLs, start, end, title, mode)

	fmt.Println("全部下载完成！")
}
