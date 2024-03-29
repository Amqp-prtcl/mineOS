package zip

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

//Zip will ignore any symlink and hardlinks are dereferenced
func ZipFile(srcFolder string, dstFile string) error {
	f, err := os.Open(dstFile)
	if err != nil {
		return err
	}
	defer f.Close()
	return Zip(srcFolder, f)
}

// Zip does NOT close wr !
func Zip(srcFolder string, wr io.WriteCloser) error {
	srcFolder, err := filepath.Abs(srcFolder)
	if err != nil {
		return err
	}
	info, err := os.Lstat(srcFolder)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%v is not a folder", srcFolder)
	}

	dst := zip.NewWriter(wr)

	err = filepath.WalkDir(srcFolder, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcFolder, path)
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			fmt.Printf("ignoring symLink %s...\n", path)
			return nil
		}
		if d.IsDir() {
			_, err = dst.CreateHeader(&zip.FileHeader{
				Name:     rel + "/",
				Method:   zip.Store,
				Modified: info.ModTime(),
			})
			return err
		}

		//fmt.Printf("got file: %v\n", rel)
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		wr, err := dst.CreateHeader(&zip.FileHeader{
			Name:               rel,
			Method:             zip.Deflate,
			Modified:           info.ModTime(),
			UncompressedSize64: uint64(info.Size()),
		})
		if err != nil {
			return err
		}
		_, err = io.Copy(wr, f)
		return err
	})
	if err != nil {
		dst.Close()
		return err
	}

	return dst.Close()
}
