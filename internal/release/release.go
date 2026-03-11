package release

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/six2dez/OneListForAll/internal/config"
)

func WriteChecksums(outputDir string, files []string, outFile string) error {
	f, err := os.Create(filepath.Join(outputDir, outFile))
	if err != nil {
		return fmt.Errorf("create checksum file: %w", err)
	}
	defer f.Close()

	for _, name := range files {
		if err := ensureOutputFile(outputDir, name); err != nil {
			return err
		}
		full := filepath.Join(outputDir, name)
		h, err := fileSHA256(full)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(f, "%s  %s\n", h, name); err != nil {
			return fmt.Errorf("write checksum: %w", err)
		}
	}
	return nil
}

func Package7z(outputDir string, cfg config.Release) error {
	if !cfg.Split7z.Enabled {
		return nil
	}
	if _, err := exec.LookPath("7z"); err != nil {
		return fmt.Errorf("7z not found in PATH")
	}
	for _, out := range cfg.Outputs {
		if err := ensureOutputFile(outputDir, out); err != nil {
			return err
		}
		src := filepath.Join(outputDir, out)
		archive := filepath.Join(outputDir, out+".7z")
		vol := fmt.Sprintf("-v%dm", cfg.Split7z.VolumeMB)
		cmd := exec.Command("7z", "a", "-t7z", vol, archive, src)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("package %s: %w", out, err)
		}
	}
	return nil
}

func ensureOutputFile(outputDir, name string) error {
	inOutput := filepath.Join(outputDir, name)
	if _, err := os.Stat(inOutput); err == nil {
		return nil
	}
	rootFile := name
	if _, err := os.Stat(rootFile); err != nil {
		return fmt.Errorf("missing output file %q in %s and repo root", name, outputDir)
	}
	src, err := os.Open(rootFile)
	if err != nil {
		return err
	}
	defer src.Close()
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}
	dst, err := os.Create(inOutput)
	if err != nil {
		return err
	}
	defer dst.Close()
	if _, err := io.Copy(dst, src); err != nil {
		return err
	}
	return dst.Sync()
}

func VerifyChecksums(outputDir, checksumFile string) error {
	buf, err := os.ReadFile(filepath.Join(outputDir, checksumFile))
	if err != nil {
		return err
	}
	for _, line := range strings.Split(strings.TrimSpace(string(buf)), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) != 2 {
			return fmt.Errorf("invalid checksum line: %q", line)
		}
		expected := parts[0]
		actual, err := fileSHA256(filepath.Join(outputDir, parts[1]))
		if err != nil {
			return err
		}
		if actual != expected {
			return fmt.Errorf("checksum mismatch for %s", parts[1])
		}
	}
	return nil
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open %q: %w", path, err)
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("hash %q: %w", path, err)
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
