package dedupe

import (
	"bufio"
	"container/heap"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type chunkWriter struct {
	chunkSize int
	tempDir   string
	buffer    []string
	files     []string
}

func NewChunkWriter(chunkSize int, tempDir string) *chunkWriter {
	return &chunkWriter{chunkSize: chunkSize, tempDir: tempDir}
}

func (c *chunkWriter) Add(line string) error {
	c.buffer = append(c.buffer, line)
	if len(c.buffer) >= c.chunkSize {
		return c.flush()
	}
	return nil
}

func (c *chunkWriter) Close() ([]string, error) {
	if len(c.buffer) > 0 {
		if err := c.flush(); err != nil {
			return nil, err
		}
	}
	return c.files, nil
}

func (c *chunkWriter) flush() error {
	sort.Strings(c.buffer)
	uniq := c.buffer[:0]
	var prev string
	for i, s := range c.buffer {
		if i == 0 || s != prev {
			uniq = append(uniq, s)
			prev = s
		}
	}

	f, err := os.CreateTemp(c.tempDir, "olfa-chunk-*.txt")
	if err != nil {
		return fmt.Errorf("create chunk temp file: %w", err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for _, s := range uniq {
		if _, err := w.WriteString(s + "\n"); err != nil {
			return fmt.Errorf("write chunk: %w", err)
		}
	}
	if err := w.Flush(); err != nil {
		return fmt.Errorf("flush chunk: %w", err)
	}
	c.files = append(c.files, f.Name())
	c.buffer = c.buffer[:0]
	return nil
}

type item struct {
	line    string
	src     int
	scanner *bufio.Scanner
	file    *os.File
}

type minHeap []item

func (h minHeap) Len() int           { return len(h) }
func (h minHeap) Less(i, j int) bool { return h[i].line < h[j].line }
func (h minHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *minHeap) Push(x any) { *h = append(*h, x.(item)) }

func (h *minHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

func MergeToOutput(chunks []string, output string) (int64, error) {
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return 0, fmt.Errorf("create output dir: %w", err)
	}

	out, err := os.Create(output)
	if err != nil {
		return 0, fmt.Errorf("create output file: %w", err)
	}
	defer out.Close()

	w := bufio.NewWriter(out)
	defer w.Flush()

	h := &minHeap{}
	heap.Init(h)
	items := make([]item, 0, len(chunks))

	for i, path := range chunks {
		f, err := os.Open(path)
		if err != nil {
			return 0, fmt.Errorf("open chunk %q: %w", path, err)
		}
		s := bufio.NewScanner(f)
		s.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		if s.Scan() {
			it := item{line: s.Text(), src: i, scanner: s, file: f}
			heap.Push(h, it)
			items = append(items, it)
		} else {
			f.Close()
		}
	}

	var last string
	var written int64
	var haveLast bool

	for h.Len() > 0 {
		cur := heap.Pop(h).(item)
		if !haveLast || cur.line != last {
			if _, err := w.WriteString(cur.line + "\n"); err != nil {
				return 0, fmt.Errorf("write output: %w", err)
			}
			last = cur.line
			haveLast = true
			written++
		}
		if cur.scanner.Scan() {
			cur.line = cur.scanner.Text()
			heap.Push(h, cur)
		} else {
			if err := cur.scanner.Err(); err != nil {
				return 0, fmt.Errorf("scan chunk: %w", err)
			}
			cur.file.Close()
		}
	}

	for _, it := range items {
		_ = it.file.Close()
	}
	for _, c := range chunks {
		_ = os.Remove(c)
	}

	return written, nil
}
