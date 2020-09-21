package database

import (
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"
)

func initDataDirIfNotExists(dataDir string) error {
	if fileExist(getGenesisJSONFilePath(dataDir)) {
		return nil
	}

	err := os.MkdirAll(getDatabaseDirPath(dataDir), os.ModePerm)
	if err != nil {
		return err
	}

	err = writeGenesisToDisk(getGenesisJSONFilePath(dataDir))
	if err != nil {
		return err
	}

	return nil
}

func getDatabaseDirPath(dataDir string) string {
	return filepath.Join(dataDir, "database")
}

func getGenesisJSONFilePath(dataDir string) string {
	return filepath.Join(getDatabaseDirPath(dataDir), "genesis.json")
}

func getBlocksDbFilePath(dataDir string) string {
	return filepath.Join(getDatabaseDirPath(dataDir), "block.db")
}

func fileExist(filePath string) bool {
	_, err := os.Stat(filePath)
	if err != nil && os.IsNotExist(err) {
		return false
	}

	return true
}

func dirExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func writeEmptyBlocksDbToDisk(path string) error {
	return ioutil.WriteFile(path, []byte(""), os.ModePerm)
}

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
