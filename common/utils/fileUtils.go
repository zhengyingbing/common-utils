package utils

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

/**
 * @author: zhengyb
 * @desc: check if a folder
 * @date: 2025/3/19 16:25
 */
func IsDir(path string) bool {
	//获取文件信息
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}
	return stat.IsDir()
}

/**
 * @author: zhengyb
 * @desc: create a multi-level folders
 * @date: 2025/3/20 14:38
 */

func CreateDir(path string) error {
	if Exist(path) {
		Remove(path)
	}
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}
	return nil
}

func CreateFile(path string) (*os.File, error) {
	p := filepath.Dir(path)
	if _, err := os.Stat(p); err != nil {
		if err = os.MkdirAll(p, os.ModePerm); err != nil {
			return nil, err
		}
	}
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	return file, nil
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

/**
 * @author: zhengyb
 * @desc: copy and cover
 * @date: 2025/3/20 14:39
 */
func ForceCopy(src, dst string) error {
	return Copy(src, dst, true)
}

func Copy(src, dst string, isForced bool) error {
	if IsDir(src) {
		entries, err := os.ReadDir(src)
		if err != nil {
			return err
		}
		for _, item := range entries {
			err = Copy(filepath.Join(src, item.Name()), filepath.Join(dst, item.Name()), isForced)
		}
		return err
	} else {
		return copyFile(src, dst, isForced)
	}
}

func copyFile(src, dst string, isForced bool) error {
	if !Exist(dst) || isForced {
		srcFile, err := os.Open(src)
		if err != nil {
			return fmt.Errorf("can't %v", err)
		}
		//延迟关闭
		defer srcFile.Close()

		dstFile, err := CreateFile(dst)
		if err != nil {
			return fmt.Errorf("can't %v", err)
		}
		// 使用32KB的缓冲区提高性能
		//buf := make([]byte, 32*1024)
		//_, err = io.CopyBuffer(dstFile, srcFile, buf)
		_, err = io.Copy(dstFile, srcFile)
		if err != nil {
			return fmt.Errorf("can't Copy %v", err)
		}

		err = dstFile.Sync()
		if err != nil {
			return fmt.Errorf("can't Sync %v", err)
		}
		err = dstFile.Close()
		if err != nil {
			panic(err.Error())
			return err
		}
	}

	return nil
}

/**
 * @author: zhengyb
 * @desc: delete and copy
 * @date: 2025/3/20 14:40
 */
func Move(src, dst string, isForced bool) error {
	if IsDir(src) {
		//moveDir(srcFilePath, dstFilePath, isForced)
		files, err := os.ReadDir(src)
		if err != nil {
			fmt.Printf("无法读取源文件夹: %v\n", err)
			return err
		}

		if !Exist(dst) {
			CreateDir(dst)
		}
		for _, file := range files {
			Move(filepath.Join(src, file.Name()), filepath.Join(dst, file.Name()), isForced)
		}
	} else {
		dir := filepath.Dir(dst)
		if !Exist(dir) {
			CreateDir(dir)
		}
		if isForced {
			Remove(dst)
		}
		if !Exist(dst) || isForced {
			err := os.Rename(src, dst)
			return err
		}
	}
	return nil
}

func ReplaceFile(src, old, new string) error {
	//获取文件信息
	file, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !file.IsDir() {
		content, err := os.ReadFile(src)
		if err != nil {
			return err
		}

		newContent := strings.Replace(string(content), old, new, -1)
		if newContent != string(content) {
			file, err := os.Stat(src)
			err = os.WriteFile(src, []byte(newContent), file.Mode())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func ReplaceAllFiles(src, old, new string) error {
	//获取文件信息
	file, err := os.Stat(src)
	if err != nil {
		return err
	}
	if file.IsDir() {
		err := filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			newContent := strings.ReplaceAll(string(content), old, new)
			if newContent != string(content) {
				err = os.WriteFile(path, []byte(newContent), info.Mode())
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return err
		}

	} else {
		content, err := os.ReadFile(src)
		if err != nil {
			return err
		}
		newContent := strings.ReplaceAll(string(content), old, new)
		if newContent != string(content) {
			file, err := os.Stat(src)
			err = os.WriteFile(src, []byte(newContent), file.Mode())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

/**
 * @author: zhengyb
 * @desc: remove a folder
 * @date: 2025/3/20 14:41
 */
func Remove(src string) error {
	if Exist(src) {
		err := os.RemoveAll(src)
		if err != nil {
			return err
		}
	}
	return nil
}
