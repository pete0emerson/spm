package spm

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func prompt(message string, options []string, defaultOption string) string {
	reader := bufio.NewReader(os.Stdin)
	text := ""
	for {
		fmt.Printf("%s [%s] (%s) > ", message, strings.Join(options, "/"), defaultOption)
		text, _ = reader.ReadString('\n')
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)
		if stringInSlice(text, options) {
			return text
		} else {
			fmt.Printf("%s is not in [%s]\n", text, strings.Join(options, "/"))
		}
	}
}

// https://golangcode.com/download-a-file-from-a-url/
func downloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

// https://golangcode.com/unzip-files-in-go/
func unzip(src string, dest string) ([]string, error) {

	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}

// https://golangdocs.com/tar-gzip-in-golang
func untar(tarball, target string) (string, error) {
	reader, err := os.Open(tarball)
	if err != nil {
		return "", err
	}
	defer reader.Close()
	tarReader := tar.NewReader(reader)
	dir := ""

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return "", err
		}

		path := filepath.Join(target, header.Name)
		if header.Name == "pax_global_header" {
			continue
		}
		if dir == "" {
			dir = header.Name
		}
		info := header.FileInfo()
		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return "", err
			}
			continue
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return "", err
		}
		defer file.Close()
		_, err = io.Copy(file, tarReader)
		if err != nil {
			return "", err
		}
	}
	return dir, nil
}

// https://golangdocs.com/tar-gzip-in-golang
func unGzip(source, target string) error {
	reader, err := os.Open(source)
	if err != nil {
		return err
	}
	defer reader.Close()

	archive, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer archive.Close()

	target = filepath.Join(target, archive.Name)
	writer, err := os.Create(target)
	if err != nil {
		return err
	}
	defer writer.Close()

	_, err = io.Copy(writer, archive)
	return err
}

func Remove(path string, autoConfirm bool) error {
	remove := false
	if autoConfirm {
		remove = true
	} else {
		question := fmt.Sprintf("Remove %s", path)
		text := prompt(question, []string{"yes", "no"}, "no")
		if text == "yes" {
			remove = true
		}
	}
	if remove {
		log.Infof("Removing %s", path)
		err := os.RemoveAll(path)
		if err != nil {
			return err
		}
	}
	return nil
}

func Install(URI string, destinationPath string, autoConfirm bool) error {
	install := false
	if autoConfirm {
		install = true
	} else {
		question := fmt.Sprintf("Install %s to %s", URI, destinationPath)
		text := prompt(question, []string{"yes", "no"}, "no")
		if text == "yes" {
			install = true
		}

	}
	if install {

		//https://github.com/pete0emerson/tendril-example-scripts/archive/v1.0.0.tar.gz

		if strings.HasSuffix(URI, ".zip") {
			filePathArray := strings.Split(URI, "/")
			filePath := filePathArray[len(filePathArray)-1]
			err := downloadFile(filePath, URI)
			if err != nil {
				return err
			}
			files, err := unzip(filePath, destinationPath)
			if err != nil {
				return err
			}
			fmt.Println(files)
		} else if strings.HasSuffix(URI, ".tar.gz") || strings.HasSuffix(URI, ".tgz") {
			filePathArray := strings.Split(URI, "/")
			filePath := filePathArray[len(filePathArray)-1]
			err := downloadFile(filePath, URI)
			if err != nil {
				return err
			}
			ungzippedFilePath := strings.Replace(filePath, ".tar.gz", ".tar", 1)
			ungzippedFilePath = strings.Replace(ungzippedFilePath, ".tgz", ".tar", 1)
			err = unGzip(filePath, ungzippedFilePath)
			if err != nil {
				return err
			}
			err = os.Remove(filePath)
			if err != nil {
				return err
			}
			tempDir, err := untar(ungzippedFilePath, ".")
			if err != nil {
				return err
			}
			err = os.Remove(ungzippedFilePath)
			if err != nil {
				return err
			}
			err = os.Rename(tempDir, destinationPath)
			if err != nil {
				return err
			}
		} else {
			_, err := git.PlainClone(destinationPath, false, &git.CloneOptions{
				URL:      URI,
				Depth:    1,
				Progress: os.Stdout,
			})
			if err != nil {
				return err
			}
			err = os.RemoveAll(destinationPath + "/.git")
			if err != nil {
				return err
			}
		}
	}
	return nil
}
