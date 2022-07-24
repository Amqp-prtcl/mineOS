package downloads

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"mineOS/config"
	"os"
	"path/filepath"
	"time"

	"github.com/Amqp-prtcl/snowflakes"
)

var (
	ErrNoExists = errors.New("file not found")

	downloadNode = snowflakes.NewNode(3)
)

func fromIdtoTempPath(id snowflakes.ID) string {
	return filepath.Join(config.Config.DownloadFolder, "temp-"+id.String())
}

func fromIdToPath(id snowflakes.ID) string {
	return filepath.Join(config.Config.DownloadFolder, id.String())
}

func fromIdtoPathInfo(id snowflakes.ID) string {
	return filepath.Join(config.Config.DownloadFolder, id.String()+"-info.json")
}

type Info struct {
	Name            string `json:"name"`
	Size            int64  `json:"size"`
	Sha256          string `json:"sha256"`
	ExpirationStamp int64  `json:"expiration-stamp"`
}

func GetInfo(id snowflakes.ID) (*Info, error) {
	f, err := os.Open(fromIdtoPathInfo(id))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoExists
		}
		return nil, err
	}
	defer f.Close()
	var info = &Info{}
	return info, json.NewDecoder(f).Decode(&info)
}

// it is the caller's responsability to call Close
func GetFile(id snowflakes.ID) (io.ReadCloser, error) {
	f, err := os.Open(fromIdToPath(id))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoExists
		}
		return nil, err
	}
	return f, nil
}

// does no attempt at sanitizing name
func NewFile(name string, expiresIn time.Duration) (io.WriteCloser, snowflakes.ID, error) {
	id := downloadNode.NewID()

	err := os.MkdirAll(config.Config.DownloadFolder, 0666)
	if err != nil {
		fmt.Printf("Unable to create download directory: %s\n", err.Error())
		return nil, "", err
	}
	f, err := os.Create(fromIdToPath(id))
	if err != nil {
		return nil, "", err
	}

	var wr = &writer{
		filaname:        name,
		id:              id,
		expirationStamp: time.Now().Add(expiresIn).Sub(snowflakes.GetEpoch()).Milliseconds(),
		count:           0,
		hasher:          sha256.New(),
		f:               f,
	}

	return wr, id, nil
}

type writer struct {
	filaname        string
	id              snowflakes.ID
	expirationStamp int64
	count           int64
	hasher          hash.Hash
	f               *os.File
}

func (wr *writer) Write(buf []byte) (int, error) {
	n, err := wr.f.Write(buf)
	wr.count += int64(n)
	wr.hasher.Write(buf[:n])
	return n, err
}

func (wr *writer) Close() error {
	wr.f.Close()

	var info = &Info{
		Name:            wr.filaname,
		Size:            wr.count,
		Sha256:          fmt.Sprintf("%x", wr.hasher.Sum(nil)),
		ExpirationStamp: wr.expirationStamp,
	}

	i, err := wr.f.Stat()
	if err != nil {
		fmt.Printf("err: %v", err)
	}
	if i.Size() != info.Size {
		fmt.Printf("[???] sizes don't match")
	}

	f, err := os.Create(fromIdtoPathInfo(wr.id))
	if err != nil {
		return err
	}
	defer f.Close()

	err = json.NewEncoder(f).Encode(info)
	if err != nil {
		return err
	}

	return os.Rename(fromIdtoTempPath(wr.id), fromIdToPath(wr.id))
}
