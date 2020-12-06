package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

func timeago(t *time.Time) string {
	d := time.Since(*t)
	if d.Seconds() < 60 {
		seconds := int(d.Seconds())
		if seconds == 1 {
			return "1 second ago"
		}
		return fmt.Sprintf("%d seconds ago", seconds)
	} else if d.Minutes() < 60 {
		minutes := int(d.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	} else if d.Hours() < 24 {
		hours := int(d.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else {
		days := int(d.Hours()) / 24
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}

// safe
func getUserDirectory(username string) string {
	// extra filepath.clean just to be safe
	userFolder := path.Join(c.FilesDirectory, filepath.Clean(username))
	return userFolder
}

// ugh idk
func safeGetFilePath(username string, filename string) string {
	return path.Join(getUserDirectory(username), filepath.Clean(filename))
}

// TODO move into checkIfValidFile. rename it
func userHasSpace(user string, newBytes int) bool {
	userPath := path.Join(c.FilesDirectory, user)
	size, err := dirSize(userPath)
	if err != nil || size+int64(newBytes) > c.MaxUserBytes {
		return false
	}
	return true
}

func dirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}

/// Perform some checks to make sure the file is OK
func checkIfValidFile(filename string, fileBytes []byte) error {
	if len(filename) == 0 {
		return fmt.Errorf("Please enter a filename")
	}
	if len(filename) > 256 { // arbitrarily chosen
		return fmt.Errorf("Filename is too long")
	}
	ext := strings.ToLower(path.Ext(filename))
	found := false
	for _, mimetype := range c.OkExtensions {
		if ext == mimetype {
			found = true
		}
	}
	if !found {
		return fmt.Errorf("Invalid file extension: %s", ext)
	}
	if len(fileBytes) > c.MaxFileBytes {
		return fmt.Errorf("File too large. File was %d bytes, Max file size is %d", len(fileBytes), c.MaxFileBytes)
	}
	//
	return nil
}

func zipit(source string, target io.Writer) error {
	archive := zip.NewWriter(target)

	info, err := os.Stat(source)
	if err != nil {
		return nil
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if baseDir != "" {
			header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})

	archive.Close()

	return err
}
