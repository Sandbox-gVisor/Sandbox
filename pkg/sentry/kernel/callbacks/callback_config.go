package callbacks

import (
	"encoding/json"
	"errors"
	"syscall"
)

type CallbackDto struct {
	Sysno          int    `json:"sysno"`
	EntryPoint     string `json:"entry-point"`
	CallbackSource string `json:"source"`
	Type           string `json:"type"`
}

type CallbackConfigDto struct {
	SocketFileName string `json:"runtime-socket"`

	CallbackDtos []CallbackDto `json:"callbacks"`
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
