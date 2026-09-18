package file

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"remote/model"
)

// Apply 执行 UI 发起的文件系统操作（新建目录 / 删除）
func Apply(msg model.ClientFileReceive) error {
	switch msg.Op {
	case "fs_mkdir":
		return Mkdir(msg.Target)
	case "fs_delete":
		return Remove(msg.Paths)
	}
	return fmt.Errorf("未知操作: %s", msg.Op)
}

// Mkdir 新建目录（含父目录）
func Mkdir(path string) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("路径为空")
	}
	return os.MkdirAll(path, 0755)
}

// Remove 删除文件或目录（目录递归）。删除是不可恢复操作，这里挡掉根目录。
func Remove(paths []string) error {
	if len(paths) == 0 {
		return errors.New("没有选择要删除的内容")
	}
	var errs []string
	for _, p := range paths {
		if strings.TrimSpace(p) == "" {
			continue
		}
		if isRoot(p) {
			errs = append(errs, fmt.Sprintf("%s 是根目录，已跳过", p))
			continue
		}
		if err := os.RemoveAll(p); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", filepath.Base(p), err))
		}
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "；"))
	}
	return nil
}

// isRoot 判断是不是盘符根 / 文件系统根，避免误删整盘。
// 注意 filepath.Clean("C:") 在 Windows 上返回 "C:."，所以两种形态都要认。
func isRoot(p string) bool {
	s := strings.TrimSpace(p)
	if s == "" {
		return false
	}
	clean := filepath.Clean(strings.ReplaceAll(s, "/", string(filepath.Separator)))
	if clean == string(filepath.Separator) || clean == "." {
		return true
	}
	if len(clean) == 2 && clean[1] == ':' {
		return true
	}
	if len(clean) == 3 && clean[1] == ':' && (clean[2] == '.' || os.IsPathSeparator(clean[2])) {
		return true
	}
	return false
}
