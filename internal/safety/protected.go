package safety

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// protectedRel are home-relative directories clean-me must never offer or
// delete — credentials, keys, and tool state that would be catastrophic to
// lose. Matched as the path itself or any ancestor.
var protectedRel = []string{
	".ssh", ".aws", ".gnupg", ".kube", ".cargo",
	".docker", ".config/clean-me", ".config/gcloud", ".password-store",
}

// IsProtected reports whether a filesystem path must never be touched: the home
// root or "/", a hardcoded protected directory, any `.git` component, or a
// user-supplied exclude entry (exact, prefix, or glob).
func IsProtected(path string, exclude []string) bool {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = filepath.Clean(path)
	}
	if abs == "/" {
		return true
	}
	if home, err := os.UserHomeDir(); err == nil {
		if abs == filepath.Clean(home) {
			return true
		}
		for _, rel := range protectedRel {
			p := filepath.Join(home, rel)
			if abs == p || strings.HasPrefix(abs, p+string(os.PathSeparator)) {
				return true
			}
		}
	}
	// Never delete anything inside a git repo's metadata.
	for _, seg := range strings.Split(abs, string(os.PathSeparator)) {
		if seg == ".git" {
			return true
		}
	}
	for _, e := range exclude {
		e = strings.TrimSpace(e)
		if e == "" {
			continue
		}
		if abs == e || strings.HasPrefix(abs, strings.TrimRight(e, "/")+string(os.PathSeparator)) {
			return true
		}
		if ok, _ := filepath.Match(e, abs); ok {
			return true
		}
	}
	return false
}

// LoadExclude reads the user allow-list from ~/.config/clean-me/exclude (one
// path or glob per line, # comments allowed). Missing file → nil, no error.
func LoadExclude() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	f, err := os.Open(filepath.Join(home, ".config", "clean-me", "exclude"))
	if err != nil {
		return nil
	}
	defer f.Close()

	var out []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "~/") {
			line = filepath.Join(home, line[2:])
		}
		out = append(out, line)
	}
	return out
}
