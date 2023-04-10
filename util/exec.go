package util

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	ping "github.com/go-ping/ping"
)

// ErrDestinationUnreachable 目的地址不可达错误
var ErrDestinationUnreachable = errors.New("destination unreachable")

// ExecutionOptions 命令执行可选参数
type ExecutionOptions struct {
	Env     []string
	Shadows []string
	Stdin   []string
	Timeout int
}

// PingOptions ping可选参数
type PingOptions struct {
	Count    int
	Interval int
	Timeout  int
}

// Executor 命令行执行器
type Executor interface {
	// 命令执行
	Exec(opts *ExecutionOptions, cmd string, args ...string) (output []byte, err error)
	// Ping远程主机
	Ping(opts *PingOptions, host string) error
	// 设置日志实现
	SetLog(log Logger)
}

// NewBash 返回Bash实现的命令执行器
func NewBash() Executor {
	return new(Bash)
}

// Bash 本地Shell执行器
type Bash struct {
	log Logger
}

// SetLog 更改日志实现
func (bash *Bash) SetLog(log Logger) {
	bash.log = log
}

const (
	shell = "/bin/bash"
)

// Ping 主机可达性检查
func (bash *Bash) Ping(opts *PingOptions, host string) error {
	var count, interval, timeout int
	if opts != nil {
		count, interval, timeout = opts.Count, opts.Interval, opts.Timeout
	}

	pinger, err := ping.NewPinger(host)
	if err != nil {
		return fmt.Errorf("ping error: %w", err)
	}
	// defer pinger.Stop() // 内部channel可能已经关闭，重关闭channel会导致panic。

	if count > 0 {
		pinger.Count = count
	}
	if interval > 0 {
		pinger.Interval = time.Duration(interval) * time.Second
	}
	if timeout > 0 {
		pinger.Timeout = time.Duration(timeout) * time.Second
	}
	if runtime.GOOS == "windows" {
		pinger.SetPrivileged(true)
	}
	pinger.Run()
	stats := pinger.Statistics()
	if stats == nil || stats.PacketsRecv <= 0 || stats.PacketLoss == 1 {
		return ErrDestinationUnreachable
	}
	return nil
}

// Exec 执行指定命令
func (bash *Bash) Exec(opts *ExecutionOptions, cmd string, args ...string) (output []byte, err error) {
	if opts != nil && len(opts.Stdin) > 0 {
		return bash.execWithStdin(opts, cmd, args...)
	}

	var env, shadows []string
	if opts != nil {
		env, shadows = opts.Env, opts.Shadows
	}

	if bash.log != nil {
		cmdArgs := bash.setShadows(fmt.Sprintf("%s %s", cmd, strings.Join(args, " ")), shadows...)
		bash.log.Debugf("==> %s", cmdArgs)
	}

	scriptFile, err := bash.genTempScript([]byte(fmt.Sprintf("export LC_ALL=C\n%s %s", cmd, strings.Join(args, " "))))
	if err != nil {
		if bash.log != nil {
			bash.log.Error(err)
		}
		return nil, err
	}
	defer os.Remove(scriptFile)

	command := exec.Command(shell, scriptFile)
	command.Env = env
	output, err = command.CombinedOutput()
	if output != nil && bash.log != nil {
		bash.log.Debugf("\n--------------------stdout/stderr begin--------------------\n%s\n--------------------stdout/stderr end--------------------", string(output))
	}

	if err != nil && bash.log != nil {
		bash.log.Debugf(err.Error())
		err = fmt.Errorf("exec error: %s", string(output))
	}
	return output, err
}

// execWithStdin 本地命令执行并读取stdin内容后返回stdout内容。
func (bash *Bash) execWithStdin(opts *ExecutionOptions, cmd string, args ...string) (output []byte, err error) {
	var env, shadows, inputs []string
	if opts != nil {
		env, shadows, inputs = opts.Env, opts.Shadows, opts.Stdin
	}

	if bash.log != nil {
		cmdArgs := bash.setShadows(fmt.Sprintf("%s %s", cmd, strings.Join(args, " ")), shadows...)
		bash.log.Debugf("==> %s", cmdArgs)
	}

	command := exec.Command(cmd, args...)
	command.Env = env
	stdin, err := command.StdinPipe()
	if err != nil {
		return nil, err
	}

	go func() {
		defer stdin.Close()
		for i := range inputs {
			fmt.Fprintln(stdin, inputs[i])
			if i < len(inputs)-1 {
				time.Sleep(time.Second)
			}
		}
	}()

	output, err = command.CombinedOutput()
	if output != nil && bash.log != nil {
		bash.log.Debugf("\n--------------------stdout begin--------------------\n%s\n--------------------stdout end--------------------", string(output))
	}
	if err != nil && bash.log != nil {
		bash.log.Errorf("\n--------------------stderr begin--------------------\n%s\n--------------------stderr end--------------------", err.Error())
		return nil, err
	}
	return output, err
}

// genTempScript 在系统临时目录生成可执行脚本文件
func (bash *Bash) genTempScript(content []byte) (scriptFile string, err error) {
	scriptFile = filepath.Join(os.TempDir(), fmt.Sprintf("%d.sh", time.Now().UnixNano())) // TODO 并发bug
	if err = ioutil.WriteFile(scriptFile, content, 0744); err != nil {
		return "", err
	}
	return scriptFile, nil
}

func (bash *Bash) setShadows(src string, shadows ...string) string {
	for i := range shadows {
		if shadows[i] == "" {
			continue
		}
		src = strings.ReplaceAll(src, shadows[i], "***")
	}
	return src
}
