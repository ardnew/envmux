package builtin

import (
	"os"
	"path/filepath"

	"github.com/ardnew/mung"
)

func cwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		return pathAbs(".")
	}

	return cwd
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

func filePerm(path string) uint32 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}

	return uint32(info.Mode().Perm())
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

func mungPrefix(key string, prefix ...string) string {
	return mung.Make(
		mung.WithSubjectItems(key),
		mung.WithDelim(string(os.PathListSeparator)),
		mung.WithPrefixItems(prefix...),
	).String()
}

func mungPrefixIf(
	key string,
	predicate func(string) bool,
	prefix ...string,
) string {
	return mung.Make(
		mung.WithSubjectItems(key),
		mung.WithDelim(string(os.PathListSeparator)),
		mung.WithPrefixItems(prefix...),
		mung.WithPredicate(predicate),
	).String()
}
