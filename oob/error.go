package oob

import (
	"errors"
	"fmt"
)

var (
	// ErrChannelNotFound impi channel未发现
	ErrChannelNotFound = errors.New("channel not found")
	// ErrUnknownHardware 未知的OOB硬件类型
	ErrUnknownHardware = errors.New("unknown OOB hardware")
	// ErrOOBIPAndSNUnmatched 带外IP和设备序列号不匹配
	ErrOOBIPAndSNUnmatched = errors.New("oob ip and sn do not match")
)

// UserNotFoundError 用户不存在错误
type UserNotFoundError struct {
	Name string // 带外用户名
}

// NewUserNotFoundError 返回用户不存在错误实例
func NewUserNotFoundError(name string) error {
	return &UserNotFoundError{
		Name: name,
	}
}

func (e *UserNotFoundError) Error() string {
	return fmt.Sprintf("%q user does not exist", e.Name)
}

// IsUserNotFoundError 判断是否是用户不存在错误，若是返回true，反之为false。
func IsUserNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*UserNotFoundError)
	return ok
}

// IPUnreachableError 带外IP不可达错误
type IPUnreachableError struct {
	ip  string
	err error
}

// NewIPUnreachableError 返回带外IP不可达错误
func NewIPUnreachableError(ip string, err error) error {
	return &IPUnreachableError{
		ip:  ip,
		err: err,
	}
}

// Error 返回错误描述
func (e *IPUnreachableError) Error() string {
	return fmt.Sprintf("OOB IP %q is unreachable: %s", e.ip, e.err.Error())
}

// IsIPUnreachableError 判断是否是IP不可达错误，若是返回true，反之为false。
func IsIPUnreachableError(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*IPUnreachableError)
	return ok
}

// FRUDeviceNotPresentError FRU设备不存在错误
type FRUDeviceNotPresentError struct {
	id  string // fru device id
	err error
}

// NewFRUDeviceNotPresentError 返回FRU设备不存在错误实例
func NewFRUDeviceNotPresentError(fruDeviceID string, err error) error {
	return &FRUDeviceNotPresentError{
		id:  fruDeviceID,
		err: err,
	}
}

// Error 返回错误描述
func (e *FRUDeviceNotPresentError) Error() string {
	return fmt.Sprintf("ipmitool fru list %s error: %s", e.id, e.err)
}

// IsFRUDeviceNotPresentError 判断是否是FRU设备不存在错误
func IsFRUDeviceNotPresentError(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*FRUDeviceNotPresentError)
	return ok
}

// UsernamePasswordError 带外用户名、密码不匹配错误
type UsernamePasswordError struct {
	err error
}

// NewUsernamePasswordError 返回带外用户名、密码不匹配错误
func NewUsernamePasswordError(err error) error {
	return &UsernamePasswordError{
		err: err,
	}
}

func (e *UsernamePasswordError) Error() string {
	return fmt.Sprintf("username and password do not match: %s", e.err.Error())
}

// IsUsernamePasswordError 判断是否是用户名、密码不匹配错误，若是返回true，反之为false。
func IsUsernamePasswordError(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*UsernamePasswordError)
	return ok
}
