package file

import (
	"hash/crc32"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"remote/model"
)

func TestPercent(t *testing.T) {
	cases := []struct {
		done, total int64
		want        int
	}{
		{0, 0, 0},
		{0, 100, 0},
		{50, 100, 50},
		{99, 100, 99},
		{100, 100, 99}, // 未收到最终状态前不显示 100
		{200, 100, 99},
	}
	for _, c := range cases {
		if got := percent(c.done, c.total); got != c.want {
			t.Errorf("percent(%d, %d) = %d, want %d", c.done, c.total, got, c.want)
		}
	}
}

func TestFileCrc(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "a.bin")
	data := make([]byte, 1000)
	for i := range data {
		data[i] = byte(i % 251)
	}
	if err := os.WriteFile(p, data, 0644); err != nil {
		t.Fatal(err)
	}

	whole, err := fileCrc(p, 0)
	if err != nil {
		t.Fatal(err)
	}
	if want := crc32.Checksum(data, crcTable); whole != want {
		t.Fatalf("整文件 CRC = %d, want %d", whole, want)
	}

	// n > 0 只算前 n 字节
	prefix, err := fileCrc(p, 400)
	if err != nil {
		t.Fatal(err)
	}
	if want := crc32.Checksum(data[:400], crcTable); prefix != want {
		t.Fatalf("前缀 CRC = %d, want %d", prefix, want)
	}

	// 内容被改过（长度不变）→ 前缀 CRC 必须变，这正是续传不能只比大小的原因
	altered := append([]byte(nil), data...)
	altered[10] ^= 0xFF
	if err := os.WriteFile(p, altered, 0644); err != nil {
		t.Fatal(err)
	}
	changed, err := fileCrc(p, 400)
	if err != nil {
		t.Fatal(err)
	}
	if changed == prefix {
		t.Fatal("内容被改过后前缀 CRC 不应相同")
	}
	if info, _ := os.Stat(p); info.Size() != int64(len(data)) {
		t.Fatal("该用例要求文件总大小不变")
	}
}

func TestResumeOffset(t *testing.T) {
	cases := []struct {
		name                string
		offset              int64
		ackedCrc, prefixCrc uint32
		want                int64
	}{
		{"CRC 一致，可续传", 500, 12345, 12345, 500},
		{"CRC 不一致，从头传", 500, 12345, 999, 0},
		{"偏移为 0，没有可续内容", 0, 12345, 12345, 0},
		{"负数偏移保险起见从头传", -1, 1, 1, 0},
	}
	for _, c := range cases {
		if got := resumeOffset(c.offset, c.ackedCrc, c.prefixCrc); got != c.want {
			t.Errorf("%s: resumeOffset(%d,%d,%d) = %d, want %d",
				c.name, c.offset, c.ackedCrc, c.prefixCrc, got, c.want)
		}
	}
}

func TestIsRoot(t *testing.T) {
	cases := map[string]bool{
		`C:\`:      true,
		`C:`:       true,
		`C:/`:      true,
		`/`:        true,
		`.`:        true,
		`C:\Users`: false,
		`D:\data`:  false,
		`/home`:    false,
	}
	for p, want := range cases {
		if got := isRoot(p); got != want {
			t.Errorf("isRoot(%q) = %v, want %v", p, got, want)
		}
	}
}

func TestMkdirAndRemove(t *testing.T) {
	root := t.TempDir()

	deep := filepath.Join(root, "a", "b", "c")
	if err := Mkdir(deep); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(deep); err != nil || !info.IsDir() {
		t.Fatalf("目录未创建: %v", err)
	}

	if err := os.WriteFile(filepath.Join(deep, "x.txt"), []byte("hi"), 0644); err != nil {
		t.Fatal(err)
	}
	// 删目录必须连同内容一起删掉
	if err := Remove([]string{deep}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(deep); !os.IsNotExist(err) {
		t.Fatal("目录应被递归删除")
	}

	// 不存在的路径按成功处理
	if err := Remove([]string{filepath.Join(root, "nope")}); err != nil {
		t.Fatalf("删除不存在的路径不应报错: %v", err)
	}

	// 空列表要拦住
	if err := Remove(nil); err == nil {
		t.Fatal("空列表应报错")
	}

	// 根目录必须拒绝（这里只是断言返回错误，实际不会删任何东西）
	if err := Remove([]string{`C:\`}); err == nil {
		t.Fatal("根目录应被拒绝")
	}

	if err := Mkdir(""); err == nil {
		t.Fatal("空路径应报错")
	}
}

func TestApplyDispatch(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "made")

	if err := Apply(model.ClientFileReceive{Op: "fs_mkdir", Target: dir}); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		t.Fatal("fs_mkdir 没有建出目录")
	}
	if err := Apply(model.ClientFileReceive{Op: "fs_delete", Paths: []string{dir}}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatal("fs_delete 没有删掉目录")
	}
	if err := Apply(model.ClientFileReceive{Op: "没这个操作"}); err == nil {
		t.Fatal("未知操作应报错")
	}
}

func TestPrepareFileResume(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "sub", "a.bin")

	// 文件不存在：应创建（含父目录）并从 0 开始
	if off := prepareFile(p, 100); off != 0 {
		t.Fatalf("不存在的文件 offset = %d, want 0", off)
	}
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("文件未创建: %v", err)
	}

	// 只写了一部分：应返回已有字节数（续传）
	if err := os.WriteFile(p, make([]byte, 40), 0644); err != nil {
		t.Fatal(err)
	}
	if off := prepareFile(p, 100); off != 40 {
		t.Fatalf("不完整文件 offset = %d, want 40", off)
	}

	// 已完整：返回完整大小
	if err := os.WriteFile(p, make([]byte, 100), 0644); err != nil {
		t.Fatal(err)
	}
	if off := prepareFile(p, 100); off != 100 {
		t.Fatalf("完整文件 offset = %d, want 100", off)
	}

	// 比预期大：视为损坏，清空重来
	if err := os.WriteFile(p, make([]byte, 150), 0644); err != nil {
		t.Fatal(err)
	}
	if off := prepareFile(p, 100); off != 0 {
		t.Fatalf("超大文件 offset = %d, want 0", off)
	}
	info, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() != 0 {
		t.Fatalf("超大文件未清空, size = %d", info.Size())
	}
}

func TestCollectDir(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "sub", "deep"), 0755); err != nil {
		t.Fatal(err)
	}
	write := func(rel string, n int) {
		if err := os.WriteFile(filepath.Join(root, rel), make([]byte, n), 0644); err != nil {
			t.Fatal(err)
		}
	}
	write("a.txt", 10)
	write(filepath.Join("sub", "b.txt"), 20)
	write(filepath.Join("sub", "deep", "c.txt"), 30)

	dirs, files, total := collectDir(root, "/dest")
	if total != 60 {
		t.Fatalf("total = %d, want 60", total)
	}
	if len(files) != 3 {
		t.Fatalf("files = %d, want 3", len(files))
	}
	if len(dirs) != 2 {
		t.Fatalf("dirs = %d, want 2", len(dirs))
	}
	// 父目录必须排在子目录之前，接收端才能按顺序建目录
	if got := filepath.ToSlash(dirs[0]); !strings.HasSuffix(got, "/sub") {
		t.Fatalf("dirs[0] = %s, want 以 /sub 结尾", got)
	}
	if got := filepath.ToSlash(dirs[1]); !strings.HasSuffix(got, "/sub/deep") {
		t.Fatalf("dirs[1] = %s, want 以 /sub/deep 结尾", got)
	}
	// 每个文件的落盘路径必须带正确的相对结构
	got := make([]string, 0, len(files))
	for _, f := range files {
		got = append(got, filepath.ToSlash(f.dest))
	}
	sort.Strings(got)
	want := []string{"/dest/a.txt", "/dest/sub/b.txt", "/dest/sub/deep/c.txt"}
	sort.Strings(want)
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("dest[%d] = %s, want %s (全部: %v)", i, got[i], want[i], got)
		}
	}
}
