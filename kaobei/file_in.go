package kaobei

import (
	"os"
	"path/filepath"
	"strings"
)

func MoveImagesToSubfolders(mainFolder string) {
	files, err := os.ReadDir(mainFolder)
	if err != nil {
		return
	}
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		name := f.Name()
		if strings.HasSuffix(name, ".webp") || strings.HasSuffix(name, ".jpg") || strings.HasSuffix(name, ".png") {
			parts := strings.Split(name, "_")
			if len(parts) > 0 {
				subfolderNumber := parts[0]
				subfolderPath := filepath.Join(mainFolder, subfolderNumber)
				os.MkdirAll(subfolderPath, os.ModePerm)
				srcPath := filepath.Join(mainFolder, name)
				destPath := filepath.Join(subfolderPath, name)
				os.Rename(srcPath, destPath)
			}
		}
	}
}
