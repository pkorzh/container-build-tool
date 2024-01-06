package json

import (
	"encoding/json"
	"os"
)

func ParseJSON[T any](path string) (*T, error) {
	var contents, err = os.Open(path)
	if err != nil {
		return nil, err
	}
	defer contents.Close()

	data := new(T)
	if err := json.NewDecoder(contents).Decode(data); err != nil {
		return nil, err
	}

	return data, nil
}
