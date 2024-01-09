package filediff

import (
	"bytes"
	b64 "encoding/base64"
	"os"
)

func decodeFile(fileContents string) ([]byte, error) {
	decodedContents, err := b64.StdEncoding.DecodeString(fileContents)
	if err != nil {
		return []byte{}, err
	}

	return decodedContents, nil
}

func IsSameFile(fileName string, encodedContent string) (bool, error) {
	decodedContent, err := decodeFile(encodedContent)
	if err != nil {
		return true, err
	}

	fileContents, err := os.ReadFile(fileName)
	if err != nil {
		return true, err
	}

	return bytes.Equal(fileContents, decodedContent), nil
}

func ReplaceFile(fileName string, encodedFileContents string) error {
	fileContents, err := decodeFile(encodedFileContents)
	if err != nil {
		return err
	}

	if err := os.WriteFile(fileName, fileContents, 0600); err != nil {
		return err
	}

	return nil
}
