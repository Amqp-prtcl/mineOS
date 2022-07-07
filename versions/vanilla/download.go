package vanilla

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type VersionMeta struct {
	Downloads struct {
		Server File `json:"server"`
	} `json:"downloads"`
}

type File struct {
	Sha1 string `json:"sha1"`
	Size int64  `json:"size"`
	Url  string `json:"url"`
}

func (v *Version) DownloadServer(path string) error {
	fmt.Printf("downloading meta json\n")
	mf, err := getDownloadFileFromVersionUrl(v.URL)
	if err != nil {
		return err
	}
	fmt.Printf("downloading server jar\n")
	err = downloadFile(path, mf.Url)
	if err != nil {
		return err
	}

	fmt.Printf("checking size\n")
	stat, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if stat.IsDir() { // should never trigger
		return fmt.Errorf("downloaded file is a directory ??? (should not be possible, if you see this please immediately report it to software developers)")
	}
	if stat.Size() != mf.Size {
		return fmt.Errorf("invalid download size: expecting %v but got %v bytes", mf.Size, stat.Size())
	}

	fmt.Printf("verifying file checkdum\n")
	sum, err := Sha1Sum(path)
	if err != nil {
		return fmt.Errorf("failed to compute file hash")
	}
	if sum != mf.Sha1 {
		return fmt.Errorf("invalid sha1: execting -> %q but got -> %q", mf.Sha1, sum)
	}
	return nil
}

func getDownloadFileFromVersionUrl(vrsUrl string) (File, error) {
	var meta = VersionMeta{}
	resp, err := http.Get(vrsUrl)
	if err != nil {
		return meta.Downloads.Server, err
	}
	err = json.NewDecoder(resp.Body).Decode(&meta)
	return meta.Downloads.Server, err
}

func downloadFile(filepath string, url string) (err error) {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}
	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func Sha1Sum(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
