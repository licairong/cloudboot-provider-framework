package oob

import (
	"github.com/licairong/cloudboot-provider-framework/util"
)

const (
	// LANInterface IPMI v1.5 LAN Interface
	LANInterface = "lan"
	// LANPlusInterface IPMI v2.0 RMCP+ LAN Interface
	LANPlusInterface = "lanplus"
)

// Options 选项
type Options struct {
	Interface string            // 接口(协议)
	Hostname  string            // 带外主机名/IP
	Username  string            // 带外用户名
	Password  string            // 带外密码
	ChannelID int               // 通道ID
	Debug     bool              // 若开启debug，会将关键日志信息写入console。
	Log       util.Logger   // 日志实例
	Executor  util.Executor // 执行器实例
}

// WithRemote 设置IPMI远程操作参数
func WithRemote(i, hostname, username, password string) func(*Options) {
	return func(opts *Options) {
		opts.Interface = i
		opts.Hostname = hostname
		opts.Username = username
		opts.Password = password
	}
}

// WithLog 设置日志实例
func WithLog(log util.Logger) func(*Options) {
	return func(opts *Options) {
		opts.Log = log
	}
}

// WithExecutor 设置执行器实例
func WithExecutor(executor util.Executor) func(*Options) {
	return func(opts *Options) {
		opts.Executor = executor
	}
}

// WithDebug 设置带外debug模式
func WithDebug(debug bool) func(*Options) {
	return func(opts *Options) {
		opts.Debug = debug
	}
}

// WithChannelID 设置channel id
func WithChannelID(id int) func(*Options) {
	return func(opts *Options) {
		opts.ChannelID = id
	}
}
