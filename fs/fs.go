package fs

import (
	"os"
	"os/user"
	"path"
	"strings"
)

// ExpandPath will expands a file path
// 1. replace tilde with users home dir
// 2. expands embedded environment variables
// 3. cleans the path, e.g. /a/b/../c -> /a/c
func ExpandPath(p string) string {
	i := strings.Index(p, ":")
	if i > 0 {
		return p
	}

	i = strings.Index(p, "@")
	if i > 0 {
		return p
	}

	if strings.HasPrefix(p, "~/") || strings.HasPrefix(p, "~\\") {
		home := homeDir()
		if home != "" {
			p = home + p[1:]
		}
	}

	return path.Clean(os.ExpandEnv(p))
}

func homeDir() string {
	home := os.Getenv("HOME")
	if home == "" {
		return ""
	}

	usr, err := user.Current()
	if err != nil {
		return ""
	}

	return usr.HomeDir
}
