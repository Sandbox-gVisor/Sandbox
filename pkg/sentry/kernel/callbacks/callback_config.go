package callbacks

import (
	"encoding/json"
	"errors"
	"syscall"
)

type CallbackDto struct {
	Sysno          int    `json:"sysno"`
	EntryPoint     string `json:"entry-point"`
	CallbackSource string `json:"callback-source"`
	Type           string `json:"type"`
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

func Parse(configFD int) ([]CallbackDto, error) {

	var data []byte
	if err := readAllBytes(configFD, &data); err != nil {
		return nil, err
	}

	var callbacks []CallbackDto
	if err := json.Unmarshal(data, &callbacks); err != nil {
		return nil, err
	}

	return callbacks, nil
}
