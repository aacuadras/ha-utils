package filediff

import (
	"bytes"
	b64 "encoding/base64"
	"os"
)

// This function decodes the contents of a base64 encoded file
func decodeFile(fileContents string) ([]byte, error) {
	decodedContents, err := b64.StdEncoding.DecodeString(fileContents)
	if err != nil {
		return []byte{}, err
	}

	return decodedContents, nil
}

// This function checks if the contents of a file match the contents of the file currently in the system
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

// This function replaces the contents of an existing file
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

// This function creates a test file to run integration tests against grpc operations
func CreateTestFile(content string, fileName string) error {
	// If test directory does not exist, create it
	if _, err := os.Stat("./test_files"); os.IsNotExist(err) {
		if err := os.Mkdir("test_files/", 0700); err != nil {
			return err
		}
	}
	// Hardcode test_files directory since this method is only intended to be used with tests
	if err := os.WriteFile("./test_files/"+fileName, []byte(content), 0755); err != nil {
		return err
	}

	return nil
}
