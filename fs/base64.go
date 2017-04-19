package fs

import (
	"encoding/base64"
	"io/ioutil"
)

// Base64StrFromFile : Encode the file into the base64 string
func Base64StrFromFile(path string) (string, error) {
	buff, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buff), nil

}

// FileFromBase64Str : Decode the base64 string into a file
func FileFromBase64Str(code string, dest string) error {
	buff, err := base64.StdEncoding.DecodeString(code)

	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(dest, buff, 07400); err != nil {
		return err
	}

	return nil
}
