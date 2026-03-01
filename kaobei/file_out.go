package kaobei

import (
	"os"
	"path/filepath"
)

func MoveImagesToMainFolder(mainFolder string) {
	entries, err := os.ReadDir(mainFolder)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if entry.IsDir() {
			subfolder := filepath.Join(mainFolder, entry.Name())
			subfiles, err := os.ReadDir(subfolder)
			if err != nil {
				continue
			}
			for _, subf := range subfiles {
				if !subf.IsDir() {
					src := filepath.Join(subfolder, subf.Name())
					dest := filepath.Join(mainFolder, subf.Name())
					os.Rename(src, dest)
				}
			}
		}
	}
}
