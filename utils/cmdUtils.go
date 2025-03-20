package utils

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
	"runtime"
	"syscall"
)

type OsType string

const (
	MACOS   OsType = "mac"
	WINDOWS OsType = "windows"
	LINUX   OsType = "linux"
)

// 定义只写通道，保存打包日志
var logQueue chan<- string

func CurrentOsType() OsType {
	os := runtime.GOOS
	if os == "darwin" {
		return MACOS
	} else if os == "linux" {
		return LINUX
	}
	return WINDOWS
}

func Space() string {
	if CurrentOsType() == MACOS {
		return "/"
	}
	return "\\"
}

func ConfigLogQueue(queue chan<- string) {
	logQueue = queue
}

func ExecuteShellSync(shell string) ([]string, error) {
	log.Println("execute:  " + shell)

	result := make([]string, 0)

	//定义指针变量
	var cmd *exec.Cmd

	if CurrentOsType() == WINDOWS {
		cmd = exec.Command("cmd", "/c", shell)

	} else {
		cmd = exec.Command("/bin/bash", "-c", shell)
	}
	//设置运行命令时隐藏命令窗口
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}

	//设置实时读取命令标准输出
	//cms.Output:等待命令执行完成，一次性返回所有输出
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		pushLogQueue(err.Error())
		log.Println("err:  " + err.Error())
	}

	//设置实时读取命令错误输出
	errout, err := cmd.StderrPipe()
	go handleErr(errout)
	if err != nil {
		pushLogQueue(err.Error())
		log.Println("err:  " + err.Error())
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		pushLogQueue(err.Error())
		log.Println("err:  " + err.Error())
		return nil, err
	}

	in := bufio.NewScanner(stdout)
	for in.Scan() {
		pushLogQueue(string(in.Bytes()))
		log.Println(string(in.Bytes()))
		result = append(result, string(in.Bytes()))
	}

	err = cmd.Wait()
	if err != nil {
		pushLogQueue(err.Error())
		log.Println(err)
		return nil, err
	}
	return result, nil
}

func ExecuteShell(shell string) error {
	log.Println("execute:  " + shell)
	//指针变量
	var cmd *exec.Cmd

	if CurrentOsType() == WINDOWS {
		cmd = exec.Command("cmd", "/c", shell)

	} else {
		cmd = exec.Command("/bin/bash", "-c", shell)
	}
	//设置运行命令时隐藏命令窗口
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}

	//设置实时读取命令标准输出
	//cms.Output:等待命令执行完成，一次性返回所有输出
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		pushLogQueue(err.Error())
		log.Println("err:  " + err.Error())
	}

	//设置实时读取命令错误输出
	errout, err := cmd.StderrPipe()
	go handleErr(errout)
	if err != nil {
		pushLogQueue(err.Error())
		log.Println("err:  " + err.Error())
		return err
	}

	err = cmd.Start()
	if err != nil {
		pushLogQueue(err.Error())
		log.Println("err:  " + err.Error())
		return err
	}

	in := bufio.NewScanner(stdout)
	for in.Scan() {
		pushLogQueue(string(in.Bytes()))
		log.Println(string(in.Bytes()))

	}

	err = cmd.Wait()
	if err != nil {
		pushLogQueue(err.Error())
		log.Println(err)
		return err
	}
	return nil
}

func handleErr(errout io.ReadCloser) {
	in := bufio.NewScanner(errout)
	for in.Scan() {
		pushLogQueue(string(in.Bytes()))
		fmt.Errorf(string(in.Bytes()))
	}
}

func pushLogQueue(data string) {
	if logQueue != nil {
		logQueue <- data
	}
}
