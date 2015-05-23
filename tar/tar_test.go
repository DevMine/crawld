// Copyright 2014-2015 The DevMine authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tar

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var (
	dirPath     = "./testdata/parent"
	archivePath = dirPath + ".tar"
	tarArchive  = "./testdata/tar-archive"
	fooIOPath   = "./testdata/foo.io"

	parentPath      = dirPath
	barPath         = parentPath + "/bar.txt"
	foodirPath      = parentPath + "/foodir"
	bardirPath      = foodirPath + "/bardir"
	bazPath         = bardirPath + "/baz.txt"
	someContentPath = foodirPath + "/some-content.txt"
	symlinkDirPath  = parentPath + "/symlink-dir"
	symlinkFilePath = parentPath + "/symlink-file"
)

func TestCreateExtract(t *testing.T) {
	if err := Create(archivePath, dirPath); err != nil {
		t.Fatal(err)
	}

	if err := os.RemoveAll(dirPath); err != nil {
		t.Fatal(err)
	}

	if err := Extract(filepath.Dir(dirPath), archivePath); err != nil {
		t.Fatal(err)
	}

	if err := os.Remove(archivePath); err != nil {
		t.Error(err)
	}

	if err := testFiles(); err != nil {
		t.Error(err)
	}
}

func TestCreateExtractInPlace(t *testing.T) {
	if err := CreateInPlace(fooIOPath); err != nil {
		t.Fatal(err)
	}

	if err := ExtractInPlace(fooIOPath + ".tar"); err != nil {
		t.Fatal(err)
	}

	if err := CreateInPlace(dirPath + ".tar"); err == nil {
		t.Fatal(errors.New("error expected when given a directory path with an extension"))
	}

	if err := CreateInPlace(dirPath); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(dirPath); err == nil {
		t.Error(errors.New("directory not removed: " + dirPath))
	}

	if err := ExtractInPlace(archivePath); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(archivePath); err == nil {
		t.Error(errors.New("archive not removed: " + archivePath))
	}

	if err := ExtractInPlace(tarArchive); err == nil {
		t.Fatal(errors.New("error expected when given a tar archive without the .tar extension"))
	}
}

func testFiles() error {
	var err error
	stat := func(path string) (fi os.FileInfo) {
		if err != nil {
			return
		}
		fi, err = os.Stat(path)
		return
	}

	fiParent := stat(parentPath)
	fiFoodir := stat(foodirPath)
	fiBardir := stat(bardirPath)
	if err != nil {
		return err
	}

	if !fiParent.IsDir() {
		return errors.New("not a directory: " + parentPath)
	}

	barChecksum := "c157a79031e1c40f85931829bc5fc552"
	bar, err := ioutil.ReadFile(barPath)
	if err != nil {
		return err
	}
	expectedSum := md5.Sum(bar)
	if hex.EncodeToString(expectedSum[:]) != barChecksum {
		return errors.New("checksum mismatch for file: " + barPath)
	}

	if !fiFoodir.IsDir() {
		return errors.New("not a directory: " + foodirPath)
	}

	if !fiBardir.IsDir() {
		return errors.New("not a directory: " + bardirPath)
	}

	dest, err := filepath.EvalSymlinks(bazPath)
	if err != nil {
		return err
	}
	if dest != filepath.Clean(barPath) {
		return errors.New("symlink does not point to correct file:\n   actual => " +
			dest + "\n expected => " + filepath.Clean(barPath))
	}

	someContentChecksum := "258622b1688250cb619f3c9ccaefb7eb"
	someContent, err := ioutil.ReadFile(someContentPath)
	if err != nil {
		return err
	}
	expectedSum = md5.Sum(someContent)
	if hex.EncodeToString(expectedSum[:]) != someContentChecksum {
		return errors.New("checksum mismatch for file: " + someContentPath)
	}

	dest, err = filepath.EvalSymlinks(symlinkDirPath)
	if err != nil {
		return err
	}
	if dest != filepath.Clean(foodirPath) {
		return errors.New("symlink does not point to correct file:\n   actual => " +
			dest + "\n expected => " + filepath.Clean(foodirPath))
	}

	dest, err = filepath.EvalSymlinks(symlinkFilePath)
	if err != nil {
		return err
	}
	if dest != filepath.Clean(barPath) {
		return errors.New("symlink does not point to correct file:\n   actual => " +
			dest + "\n expected => " + filepath.Clean(barPath))
	}

	return nil
}
