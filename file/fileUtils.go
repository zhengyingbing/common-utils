package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

/**
 * @author: zhengyb
 * @desc: check if a folder
 * @date: 2025/3/19 16:25
 */
func IsDir(path string) (bool, error) {
	//获取文件信息
	stat, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return stat.IsDir(), nil
}

/**
 * @author: zhengyb
 * @desc: check file if exist
 * @date: 2025/3/19 16:53
 */
func Exist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

func Copy(src, dst string) error {
	return copyFile(src, dst, false)
}

func CopyForced(src, dst string) error {
	return copyFile(src, dst, true)
}

func copyFile(src, dst string, isForced bool) error {
	isDir, _ := IsDir(src)
	if isDir {
		entries, err := os.ReadDir(src)
		if err != nil {
			return fmt.Errorf("read file error: %v", err)
		}
		for _, item := range entries {
			err1 := copyFile(filepath.Join(src, item.Name()), filepath.Join(dst, item.Name()), isForced)
			if err1 != nil {
				return err1
			}
		}
	} else {
		p := filepath.Dir(dst)
		if _, err := os.Stat(p); err != nil {
			if err = os.MkdirAll(p, os.ModePerm); err != nil {
				return err
			}
		}

		srcFile, err := os.Open(src)
		if err != nil {
			return fmt.Errorf("can't %v", err)
		}
		defer srcFile.Close()

		if isForced || !Exist(dst) {

			dstFile, err := os.Create(dst)
			if err != nil {
				return fmt.Errorf("can't %v", err)
			}

			_, err = io.Copy(dstFile, srcFile)

			if err != nil {
				return fmt.Errorf("can't %v", err)
			}

			err = dstFile.Sync()
			if err != nil {
				return fmt.Errorf("can't %v", err)
			}
			err = dstFile.Close()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func Move(src, dst string) error {
	if Exist(dst) {
		err := os.RemoveAll(dst)
		if err != nil {
			return err
		}
	}
	err := Copy(src, dst)
	if err == nil {
		err := os.RemoveAll(src)
		if err != nil {
			return err
		}
	}
	return err
}

func Remove(src string) error {
	if Exist(src) {
		err := os.RemoveAll(src)
		if err != nil {
			return err
		}
	}
	return nil
}
