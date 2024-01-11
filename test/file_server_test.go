package test

import (
	"context"
	b64 "encoding/base64"
	"errors"
	"log"
	"net"
	"testing"

	"github.com/aacuadras/ha-utils/lib/filediff"
	"github.com/aacuadras/ha-utils/server"
	"github.com/aacuadras/ha-utils/server/pb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func createFileClient(ctx context.Context) (pb.FileUtilsClient, func()) {
	buffer := 1024 * 1024
	listener := bufconn.Listen(buffer)

	s := grpc.NewServer()
	pb.RegisterFileUtilsServer(s, server.NewFileServer())
	go func() {
		if err := s.Serve(listener); err != nil {
			log.Fatalf("error listening: %v", err)
		}
	}()

	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
		return listener.Dial()
	}), grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("error connecting to listener: %v", err)
	}

	connCloser := func() {
		err := listener.Close()
		if err != nil {
			log.Fatalf("error closing listener: %v", err)
		}

		s.Stop()
	}

	client := pb.NewFileUtilsClient(conn)
	return client, connCloser
}

func encondeFileContent(fileContent string) string {
	return b64.StdEncoding.EncodeToString([]byte(fileContent))
}

func TestCompareFile(t *testing.T) {
	ctx := context.Background()
	client, closer := createFileClient(ctx)
	defer closer()

	type expectation struct {
		output *pb.FileDiff
		err    error
	}

	testCases := map[string]struct {
		input    *pb.File
		expected expectation
	}{
		"test_same_file": {
			input: &pb.File{
				FileName:       "./test_files/test.txt",
				EncodedContent: encondeFileContent("This is a test"),
			},
			expected: expectation{
				output: &pb.FileDiff{
					IsSame: true,
				},
			},
		},
		"test_different_content": {
			input: &pb.File{
				FileName:       "./test_files/test.txt",
				EncodedContent: encondeFileContent("This is a test, but not the same one"),
			},
			expected: expectation{
				output: &pb.FileDiff{
					IsSame: false,
				},
			},
		},
		"test_nonexistent_file": {
			input: &pb.File{
				FileName:       "./test_files/nonexistenttest.txt",
				EncodedContent: encondeFileContent("Does it matter? The file doesn't exist"),
			},
			expected: expectation{
				err: errors.New("no such file or directory"),
			},
		},
		"test_non_base64": {
			input: &pb.File{
				FileName:       "./test_files/test.txt",
				EncodedContent: "Non base64 string",
			},
			expected: expectation{
				err: errors.New("illegal base64 data at input byte"),
			},
		},
	}

	for scenario, testcase := range testCases {
		t.Run(scenario, func(t *testing.T) {
			filediff.CreateTestFile("This is a test", "test.txt")
			out, err := client.CompareFile(ctx, testcase.input)

			if err != nil {
				assert.ErrorContains(t, err, testcase.expected.err.Error())
			} else {
				assert.Equal(t, testcase.expected.output.IsSame, out.IsSame)
			}
		})
	}
}

func TestSendFile(t *testing.T) {
	ctx := context.Background()
	client, closer := createFileClient(ctx)
	defer closer()

	type expectation struct {
		output *pb.ProcessedFile
	}

	testCases := map[string]struct {
		input    *pb.File
		expected expectation
	}{
		"send_different_file": {
			input: &pb.File{
				FileName:       "./test_files/test.txt",
				EncodedContent: encondeFileContent("This is a different test"),
			},
			expected: expectation{
				output: &pb.ProcessedFile{
					Processed: true,
					FileName:  "./test_files/test.txt",
				},
			},
		},
	}

	for scenario, testcase := range testCases {
		t.Run(scenario, func(t *testing.T) {
			filediff.CreateTestFile("This is a test", "test.txt")
			out, err := client.SendFile(ctx, testcase.input)
			log.Print(out.Processed)
			log.Print(testcase.expected.output.Processed)

			assert.Nil(t, err)
			assert.Equal(t, testcase.expected.output.Processed, out.Processed)
			assert.Equal(t, testcase.expected.output.FileName, out.FileName)
			assert.Equal(t, testcase.expected.output.Error, out.Error)
		})
	}
}
