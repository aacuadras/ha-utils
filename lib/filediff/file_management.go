package filediff

import (
	"bytes"
	b64 "encoding/base64"
	"os"
)

func DecodeFile(fileContents string) ([]byte, error) {
	decodedContents, err := b64.StdEncoding.DecodeString(fileContents)
	if err != nil {
		return []byte{}, err
	}

	return decodedContents, nil
}

func IsSameFile(fileName string, atpContent []byte) (bool, error) {
	fileContents, err := os.ReadFile(fileName)
	if err != nil {
		return true, err
	}

	return bytes.Equal(fileContents, atpContent), nil
}
