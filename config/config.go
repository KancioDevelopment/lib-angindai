package config

import (
	"fmt"
	"io/ioutil"
	"path"
)

func ReadConfigFile(filePath string) ([]byte, error) {
	err := validateFileExtension(filePath)
	if err != nil {
		return nil, err
	}

	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return content, nil
}

func validateFileExtension(filePath string) error {
	allowedFileExt := make(map[string]bool)
	allowedFileExt[".toml"] = true

	fileExtension := path.Ext(filePath)

	val, ok := allowedFileExt[fileExtension]
	if !(ok && val) {
		return fmt.Errorf("config: file extension %s not allowed", fileExtension)
	}

	return nil
}
