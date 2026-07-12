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
	return s.resolveExisting(path)
}

func (s scope) resolveExisting(path string) (string, error) {
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

func (s scope) resolveForWrite(path string) (string, error) {
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
	parent, err := filepath.EvalSymlinks(filepath.Dir(target))
	if err != nil {
		return "", err
	}
	target = filepath.Join(parent, filepath.Base(target))
	if target != root && !strings.HasPrefix(target, root+string(os.PathSeparator)) {
		return "", fmt.Errorf("path %q is outside workspace %q", path, root)
	}
	return target, nil
}

func (s scope) resolveForMkdir(path string) (string, error) {
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
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("path %q is outside workspace %q", path, root)
	}

	current := root
	for _, part := range strings.Split(rel, string(os.PathSeparator)) {
		if part == "." || part == "" {
			continue
		}
		current = filepath.Join(current, part)
		if _, err := os.Lstat(current); err != nil {
			if os.IsNotExist(err) {
				return target, nil
			}
			return "", err
		}
		realCurrent, err := filepath.EvalSymlinks(current)
		if err != nil {
			return "", err
		}
		if realCurrent != root && !strings.HasPrefix(realCurrent, root+string(os.PathSeparator)) {
			return "", fmt.Errorf("path %q is outside workspace %q", path, root)
		}
	}
	return target, nil
}
