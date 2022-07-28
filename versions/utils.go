package versions

import (
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
)

var (
	ErrVerIdNotFound   = fmt.Errorf("id not found")
	ErrSrvTypeNotFound = fmt.Errorf("server type not found")
)

// if size is set to -1 , it will not be checked
//
// if hash func is nil DownloadFile will not verify it
func DownloadFile(path string, url string, size int64, sum string, hash func() hash.Hash) error {
	err := unsafeDownloadFile(path, url)
	if err != nil {
		fmt.Println(3.1, err)
		return err
	}

	//verify size
	stat, err := os.Lstat(path)
	if err != nil {
		fmt.Println(3.2, err)
		os.RemoveAll(path)
		return err
	}
	if stat.IsDir() { // should never trigger
		os.RemoveAll(path)
		return fmt.Errorf("downloaded file is a directory ??? (should not be possible, if you see this please immediately report it to software developers)")
	}
	if stat.Size() != size && size != -1 {
		os.RemoveAll(path)
		return fmt.Errorf("invalid download size: expecting %v but got %v bytes", size, stat.Size())
	}

	// verify checksum
	if hash == nil {
		return nil
	}
	s, err := GetSum(path, hash)
	if err != nil {
		fmt.Println(3.3, err)
		os.RemoveAll(path)
		return fmt.Errorf("failed to compute file hash")
	}
	if s != sum {
		os.RemoveAll(path)
		return fmt.Errorf("invalid checksum: expecting -> %q but got -> %q", s, sum)
	}
	return nil
}

func unsafeDownloadFile(path string, url string) error {
	// Create the file
	err := os.MkdirAll(filepath.Dir(path), 0666)
	if err != nil {
		return err
	}
	out, err := os.Create(path)
	if err != nil {
		fmt.Println("3.1.1", err)
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
func RetrieveStructFromUrl(url string, e interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	err = json.NewDecoder(resp.Body).Decode(e)
	resp.Body.Close()
	return err
}

// return the number of fields set; v must be a pointer to a struct
//
// does not support slices; arrays; funcs; chans; or maps other than map[string]interface{}
//
// ConvertMapToStruct does NOT guarantee that all field will be set
func DecodeMapToStruct(m map[string]interface{}, v interface{}) (i int) {
	val := reflect.ValueOf(v).Elem()
	for _, field := range reflect.VisibleFields(val.Type()) {
		str, ok := field.Tag.Lookup("cache")
		if !ok {
			str = field.Name
		}
		mval, ok := m[str]
		if !ok {
			continue
		}
		vfield := val.FieldByName(field.Name)

		for vfield.Kind() == reflect.Pointer {
			vfield = vfield.Elem()
		}

		if vfield.Kind() == reflect.Struct {
			a, ok := mval.(map[string]interface{})
			if ok {
				i += DecodeMapToStruct(a, vfield.Addr().Interface())
			}
			continue
		}
		rmval := reflect.ValueOf(mval)
		if vfield.Type() == rmval.Type() {
			vfield.Set(rmval)
			i++
			continue
		}
		if rmval.CanConvert(vfield.Type()) {
			vfield.Set(rmval.Convert(vfield.Type()))
			i++
			continue
		}
		fmt.Printf("WTF ?? \n") // is slice or array or func or chan or map
	}
	return
}

// panics if s is not a struct
func EncodeStructToMap(s interface{}) map[string]interface{} {
	var m = map[string]interface{}{}
	val := reflect.ValueOf(s)
	for val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	for _, field := range reflect.VisibleFields(val.Type()) {
		tagName, ok := field.Tag.Lookup("cache")
		if !ok {
			continue
		}
		m[tagName] = val.FieldByIndex(field.Index).Interface()
	}
	return m
}
