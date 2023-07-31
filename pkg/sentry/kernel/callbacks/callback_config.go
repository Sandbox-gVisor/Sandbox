package callbacks

import (
	"encoding/json"
	"errors"
	"syscall"
)

// JsCallbackInfo contains all needed for calling user registered js callbacks
type JsCallbackInfo struct {
	// Sysno is the syscall number for which callback is registered
	Sysno int `json:"sysno"`

	// EntryPoint is the start point of execution js code
	EntryPoint string `json:"entry-point"`

	// CallbackSource is the source code of callback
	CallbackSource string `json:"source"`

	CallbackBody string `json:"body"`

	CallbackArgs []string `json:"args"`

	// Type is the callback executed before or after syscall
	Type string `json:"type"`
}

func JsCallbackInfoFromStr(str string) (*JsCallbackInfo, error) {
	bytes := []byte(str)
	info := &JsCallbackInfo{}
	err := json.Unmarshal(bytes, info)
	if err != nil {
		return nil, err
	}
	return info, nil
}

type CallbackConfigDto struct {
	// UISocket is the name for tcp socket used for communication with outside
	UISocket string `json:"runtime-socket"`

	LogSocket string `json:"log-socket"`

	CallbackDtos []JsCallbackInfo `json:"callbacks"`
}

func readAllBytes(fd int, data *[]byte) error {
	if fd < 0 {
		return errors.New("negative fd")
	}

	*data = (*data)[:0]
	buffer := make([]byte, 4096)

	for {
		n, err := syscall.Read(fd, buffer)
		if err != nil {
			return err
		}

		if n == 0 {
			break
		}

		*data = append(*data, buffer[:n]...)
	}

	return nil
}

func Parse(configFD int) (*CallbackConfigDto, error) {

	var data []byte
	if _, err := syscall.Seek(configFD, 0, 0); err != nil {
		return nil, err
	}

	if err := readAllBytes(configFD, &data); err != nil {
		return nil, err
	}

	var configDto CallbackConfigDto
	if err := json.Unmarshal(data, &configDto); err != nil {
		return nil, err
	}

	return &configDto, nil
}
