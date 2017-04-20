package fs

import (
	"container/list"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/tonychol/sink/config"
)

// GetAbsolutePath : Get the absolute path
// of the location that starts the program
func GetAbsolutePath() (string, error) {
	return os.Getwd()
}

// GetDirPathFromAgrs : Get the input path
// from the command-line arguments
func GetDirPathFromAgrs() (string, error) {
	argsArr := os.Args
	if len(argsArr) < 2 {
		err := errors.New("You should attach a file directory")
		return "", err
	}
	return os.Args[1], nil
}

// TraverseDir : A wrapper function that returns
// the filepath.WalkFunc function which would be used by filepath.Walk
func TraverseDir(fl *list.List) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
			return err
		}

		if info.IsDir() {
			fl.PushBack(path)
		}

		return nil
	}
}

// AllRecursiveDirsIn : Get all the directories string inside the dirPath
func AllRecursiveDirsIn(dirPath string) []string {
	l := list.New()

	filepath.Walk(dirPath, TraverseDir(l))

	var dirSlice = make([]string, l.Len())

	i := 0
	for e := l.Front(); e != nil; e = e.Next() {
		dirSlice[i] = e.Value.(string)
		i++
	}

	return dirSlice
}

// GetRelativeDirFromRoot returns a relative path that is lexically equivalent
// to targpath when joined to basepath (which in this case is the root dir of the synced dir.
func GetRelativeDirFromRoot(filename string) (string, error) {
	targetFileDir := getDirOfFile(filename)
	absBasePath := absPathify(config.GetInstance().SyncRoot)
	return filepath.Rel(absBasePath, targetFileDir)
}

// getFileNameFromFilePath accepts a file path and returns the name of this file.
// Reference: https://golang.org/src/path/filepath/path.go?s=12262:12291#L416
func GetFileNameFromFilePath(fpath string) string {
	return filepath.Base(fpath)
}

// GetDirOfFile returns all but the last element of path,
// typically the path's directory.
// After dropping the final element using Split, the path is Cleaned and trailing slashes are removed.
// If the path is empty, Dir returns ".".
// If the path consists entirely of slashes followed by non-slash bytes,
// Dir returns a single slash.
// In any other case, the returned path does not end in a slash.
func getDirOfFile(filepath string) string {
	return path.Dir(filepath)
}

// GetFileType : Get type according to the file info
func GetFileType(info os.FileInfo) string {
	if info.IsDir() {
		return "d"
	}
	return "f"
}

// GetFileMode : Get the fileMode string of the file
// A FileMode represents a file's mode and permission bits.
// The bits have the same definition on all systems, so that
// information about files can be moved from one system
// to another portably. Not all bits apply to all systems.
func GetFileMode(info os.FileInfo) os.FileMode {
	return info.Mode()
}

// GetCheckSumOfFile : Get the checksum string of one file
// returns error if the filePath is not valid
func GetCheckSumOfFile(filePath string) (string, error) {
	var returnMD5String string
	file, err := os.Open(filePath)
	if err != nil {
		return returnMD5String, err
	}
	defer file.Close()
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return returnMD5String, err
	}
	hashInBytes := hash.Sum(nil)[:16]
	returnMD5String = hex.EncodeToString(hashInBytes)
	return returnMD5String, nil
}

// Private functions

func absPathify(inPath string) string {
	if strings.HasPrefix(inPath, "$HOME") {
		inPath = userHomeDir() + inPath[5:]
	}

	if strings.HasPrefix(inPath, "$") {
		end := strings.Index(inPath, string(os.PathSeparator))
		inPath = os.Getenv(inPath[1:end]) + inPath[end:]
	}

	if filepath.IsAbs(inPath) {
		return filepath.Clean(inPath)
	}

	p, err := filepath.Abs(inPath)
	if err == nil {
		return filepath.Clean(p)
	}

	return ""
}

func userHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}
