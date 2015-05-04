// Copyright 2014-2015 The DevMine authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package tar provides functions to create or extract tar archives.
package tar

import (
	"archive/tar"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Create creates a tar archive from a directory.
// The resulting tar archive format is in POSIX.1 format.
func Create(destPath, dirPath string) error {
	fi, err := os.Stat(dirPath)
	if err != nil {
		return err
	}

	if !fi.IsDir() {
		return errors.New("given path is not a directory: " + dirPath)
	}

	file, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer file.Close()

	tw := tar.NewWriter(file)
	defer tw.Close()

	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		var link string
		mode := info.Mode()
		switch {
		// symlinks need special treatment
		case mode&os.ModeSymlink != 0:
			link, err = filepath.EvalSymlinks(path)
			if err != nil {
				return err
			}
			if rel, err := filepath.Rel(filepath.Dir(path), link); err == nil {
				link = rel
			}
		// we don't want to tar these sort of files
		case mode&(os.ModeNamedPipe|os.ModeSocket|os.ModeDevice) != 0:
			return nil
		}

		hdr, err := tar.FileInfoHeader(info, link)
		if err != nil {
			return err
		}
		// Name is usually only the basename when created with FileInfoHeader()
		hdr.Name, err = filepath.Rel(filepath.Dir(dirPath), path)
		if err != nil {
			return err
		}

		// Some pre-POSIX.1-1988 tar implementations indicated a directory by
		// having a trailing slash in the name. Honor that here.
		if info.IsDir() {
			hdr.Name += "/"
		}

		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}

		// no content to write if it is a directory or symlink
		if !info.Mode().IsRegular() {
			return nil
		}

		// TODO use buffer reader
		bs, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		if _, err = tw.Write(bs); err != nil {
			return err
		}

		// TODO Flush w here ?

		return nil
	})

	return err
}

// CreateInPlace creates a tar archive from a directory in place which means
// that the original directory is removed after the tar archive is created.
func CreateInPlace(destPath, dirPath string) error {
	if err := Create(destPath, dirPath); err != nil {
		return err
	}
	return os.RemoveAll(dirPath)
}

// Extract extracts a tar archive given its path.
func Extract(destPath, archivePath string) error {
	fi, err := os.Stat(archivePath)
	if err != nil {
		return err
	}

	if fi.IsDir() {
		return errors.New("given path is a directory: " + archivePath)
	}

	if err := os.MkdirAll(destPath, os.ModePerm); err != nil {
		return err
	}

	archiveFile, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer archiveFile.Close()

	tr := tar.NewReader(archiveFile)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		mode := hdr.FileInfo().Mode()
		switch {
		case mode&os.ModeDir != 0:
			if err := os.Mkdir(filepath.Join(destPath, hdr.Name), mode); err != nil {
				return err
			}
		case mode&os.ModeSymlink != 0:
			os.Symlink(hdr.Linkname, filepath.Join(destPath, hdr.Name))
		default: // consider it a regular file
			f, err := os.Create(filepath.Join(destPath, hdr.Name))
			if err != nil {
				return err
			}
			defer f.Close()

			buf := make([]byte, 8192)
			for {
				nr, err := tr.Read(buf)
				if err == io.EOF {
					break
				}
				if err != nil {
					return err
				}

				nw, err := f.Write(buf[:nr])
				if err != nil {
					return err
				}
				if nr != nw {
					return errors.New("write error: not enough (or too many) bytes written")
				}
			}
		}
	}

	return nil
}

// ExtractInPlace extracts a tar archive, in place, given its path. The
// original tar archive is removed after extraction and only its content
// remains.
func ExtractInPlace(destPath, archivePath string) error {
	if err := Extract(destPath, archivePath); err != nil {
		return err
	}
	return os.Remove(archivePath)
}
