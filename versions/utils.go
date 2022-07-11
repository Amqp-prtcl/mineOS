package versions

import (
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"net/http"
	"os"
)

// if size is set to -1 , it will not be checked
//
// if hash func is nil DownloadFile will not verify it
func DownloadFile(path string, url string, size int64, sum string, hash func() hash.Hash) error {
	err := unsafeDownloadFile(path, url)
	if err != nil {
		return err
	}

	//verify size
	stat, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if stat.IsDir() { // should never trigger
		return fmt.Errorf("downloaded file is a directory ??? (should not be possible, if you see this please immediately report it to software developers)")
	}
	if stat.Size() != size && size != -1 {
		return fmt.Errorf("invalid download size: expecting %v but got %v bytes", size, stat.Size())
	}

	// verify checksum
	if hash == nil {
		return nil
	}
	s, err := GetSum(path, hash)
	if err != nil {
		return fmt.Errorf("failed to compute file hash")
	}
	if s != sum {
		return fmt.Errorf("invalid checksum: execting -> %q but got -> %q", s, sum)
	}
	return nil
}

func unsafeDownloadFile(path string, url string) error {
	// Create the file
	out, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0777)
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
	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func GetSum(path string, hash func() hash.Hash) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := hash()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// e must be a pointer
func retreiveStructFromUrl(url string, e interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	err = json.NewDecoder(resp.Body).Decode(e)
	resp.Body.Close()
	return err
}
