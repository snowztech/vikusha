package file

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type scope struct {
	root string
}

func newScope(root string) scope {
	return scope{root: strings.TrimSpace(root)}
}

func (s scope) resolve(path string) (string, error) {
	if s.root == "" {
		return path, nil
	}

	root, err := filepath.Abs(s.root)
	if err != nil {
		return "", err
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return "", err
	}

	target := path
	if !filepath.IsAbs(target) {
		target = filepath.Join(root, target)
	}
	target, err = filepath.Abs(target)
	if err != nil {
		return "", err
	}
	target, err = filepath.EvalSymlinks(target)
	if err != nil {
		return "", err
	}
	if target != root && !strings.HasPrefix(target, root+string(os.PathSeparator)) {
		return "", fmt.Errorf("path %q is outside workspace %q", path, root)
	}
	return target, nil
}
