package web

import (
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
)

type HookProvider interface {
	HookHTML(hook string, data any) (string, bool, error)
}

func templateFiles(themeDir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(themeDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		switch strings.ToLower(filepath.Ext(path)) {
		case ".html", ".tmpl", ".gotmpl":
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, errors.New("no templates found in theme directory")
	}
	return files, nil
}
