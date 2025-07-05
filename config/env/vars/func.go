package vars

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func join(delim string, elem ...string) string {
	return strings.Join(elem, delim)
}

func envGet(key string) string {
	return os.Getenv(key)
}

func envSet(key, value string) error {
	return os.Setenv(key, value)
}

func envUnset(key string) error {
	return os.Unsetenv(key)
}

func envExists(key string) bool {
	_, ok := os.LookupEnv(key)

	return ok
}

func envIsSet(key string) bool {
	value := os.Getenv(key)
	set, err := strconv.ParseBool(value)

	return err == nil && set
}

func envPrepend(key, delim string, value ...string) error {
	current := strings.Trim(envGet(key), delim)
	prepend := strings.Trim(join(delim, value...), delim)

	if current == "" {
		return envSet(key, prepend)
	}

	return envSet(key, prepend+delim+current)
}

func envAppend(key, delim string, value ...string) error {
	current := strings.Trim(envGet(key), delim)
	appends := strings.Trim(join(delim, value...), delim)

	if current == "" {
		return envSet(key, appends)
	}

	return envSet(key, current+delim+appends)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)

	return !os.IsNotExist(err)
}

func fileIsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return info.IsDir()
}

func fileIsRegular(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return info.Mode().IsRegular()
}

func fileIsSymlink(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}

	return info.Mode()&os.ModeSymlink != 0
}

func filePerms(path string) string {
	info, err := os.Stat(path)
	if err != nil {
		return ""
	}

	return info.Mode().String()
}

func fileStat(path string) uint32 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}

	return uint32(info.Mode())
}

func pathAbs(path string) string {
	p, err := filepath.Abs(path)
	if err != nil {
		return path
	}

	return p
}

func pathCat(elem ...string) string {
	return filepath.Join(elem...)
}

func pathRel(from, to string) string {
	p, err := filepath.Rel(pathAbs(from), pathAbs(to))
	if err != nil {
		return pathCat(from, to)
	}

	return p
}
