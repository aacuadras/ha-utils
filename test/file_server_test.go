package test

import (
	"context"
	b64 "encoding/base64"
	"errors"
	"io"
	"log"
	"net"
	"os"
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

func deleteTestFiles() {
	os.RemoveAll("./test_files/")
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
			deleteTestFiles()
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
		"send_same_file": {
			input: &pb.File{
				FileName:       "./test_files/test.txt",
				EncodedContent: encondeFileContent("This is a test"),
			},
			expected: expectation{
				output: &pb.ProcessedFile{
					Processed: false,
				},
			},
		},
		"send_nonexistent_file": {
			input: &pb.File{
				FileName:       "./test_files/nonexistentfile.txt",
				EncodedContent: encondeFileContent("This is a test"),
			},
			expected: expectation{
				output: &pb.ProcessedFile{
					Processed: true,
					FileName:  "./test_files/nonexistentfile.txt",
				},
			},
		},
	}

	for scenario, testcase := range testCases {
		t.Run(scenario, func(t *testing.T) {
			filediff.CreateTestFile("This is a test", "test.txt")
			out, err := client.SendFile(ctx, testcase.input)

			assert.Nil(t, err)
			assert.Equal(t, testcase.expected.output.Processed, out.Processed)
			assert.Equal(t, testcase.expected.output.FileName, out.FileName)
			assert.Equal(t, testcase.expected.output.Error, out.Error)

			deleteTestFiles()
		})
	}
}

func TestCompareFiles(t *testing.T) {
	ctx := context.Background()
	client, closer := createFileClient(ctx)
	defer closer()

	type expectation struct {
		out []*pb.FileDiff
		err error
	}

	testcases := map[string]struct {
		input    []*pb.File
		expected expectation
	}{
		"compare_same_files": {
			input: []*pb.File{
				{
					FileName:       "./test_files/test1.txt",
					EncodedContent: encondeFileContent("This is a test"),
				},
				{
					FileName:       "./test_files/test2.txt",
					EncodedContent: encondeFileContent("This is another test"),
				},
				{
					FileName:       "./test_files/test3.txt",
					EncodedContent: encondeFileContent("This is a third and final test"),
				},
			},
			expected: expectation{
				out: []*pb.FileDiff{
					{
						IsSame: true,
					},
					{
						IsSame: true,
					},
					{
						IsSame: true,
					},
				},
			},
		},
		"compare_different_files": {
			input: []*pb.File{
				{
					FileName:       "./test_files/test1.txt",
					EncodedContent: encondeFileContent("This is a test?"),
				},
				{
					FileName:       "./test_files/test2.txt",
					EncodedContent: encondeFileContent("This is a test that should be false"),
				},
			},
			expected: expectation{
				out: []*pb.FileDiff{
					{
						IsSame: false,
					},
					{
						IsSame: false,
					},
				},
			},
		},
		"compare_mixed_files": {
			input: []*pb.File{
				{
					FileName:       "./test_files/test1.txt",
					EncodedContent: encondeFileContent("This is a test"),
				},
				{
					FileName:       "./test_files/test2.txt",
					EncodedContent: encondeFileContent("This is another test, but this should be false"),
				},
				{
					FileName:       "./test_files/test3.txt",
					EncodedContent: encondeFileContent("This is a third and final test"),
				},
			},
			expected: expectation{
				out: []*pb.FileDiff{
					{
						IsSame: true,
					},
					{
						IsSame: false,
					},
					{
						IsSame: true,
					},
				},
			},
		},
	}

	for scenario, testcase := range testcases {
		t.Run(scenario, func(t *testing.T) {
			filediff.CreateTestFile("This is a test", "test1.txt")
			filediff.CreateTestFile("This is another test", "test2.txt")
			filediff.CreateTestFile("This is a third and final test", "test3.txt")

			out, err := client.CompareFiles(ctx)

			for _, v := range testcase.input {
				if err := out.Send(v); err != nil {
					t.Errorf("Error while sending message: %v", err)
				}
			}

			if err := out.CloseSend(); err != nil {
				t.Errorf("Error closing stream: %v", err)
			}

			var outputs []*pb.FileDiff
			for {
				o, err := out.Recv()
				if errors.Is(err, io.EOF) {
					break
				}

				outputs = append(outputs, o)
			}

			if err != nil {
				assert.ErrorContains(t, err, testcase.expected.err.Error())
			} else {
				assert.Equal(t, len(testcase.expected.out), len(outputs))
				for i, o := range outputs {
					assert.Equal(t, testcase.expected.out[i].IsSame, o.IsSame)
				}
			}

			deleteTestFiles()
		})
	}
}

func TestSendFiles(t *testing.T) {
	ctx := context.Background()
	client, closer := createFileClient(ctx)
	defer closer()

	testcases := map[string]struct {
		input    []*pb.File
		expected []*pb.ProcessedFile
	}{
		"send_new_files": {
			input: []*pb.File{
				{
					FileName:       "./test_files/test1.txt",
					EncodedContent: encondeFileContent("This is a different test"),
				},
				{
					FileName:       "./test_files/test2.txt",
					EncodedContent: encondeFileContent("To see if the content of the files changes"),
				},
			},
			expected: []*pb.ProcessedFile{
				{
					Processed: true,
					FileName:  "./test_files/test1.txt",
				},
				{
					Processed: true,
					FileName:  "./test_files/test2.txt",
				},
			},
		},
		"send_content_to_same_file": {
			input: []*pb.File{
				{
					FileName:       "./test_files/test1.txt",
					EncodedContent: encondeFileContent("This is a different test"),
				},
				{
					FileName:       "./test_files/test1.txt",
					EncodedContent: encondeFileContent("To see if the content of the files changes"),
				},
			},
			expected: []*pb.ProcessedFile{
				{
					Processed: true,
					FileName:  "./test_files/test1.txt",
				},
				{
					Processed: true,
					FileName:  "./test_files/test1.txt",
				},
			},
		},
		"send_same_content_to_files": {
			input: []*pb.File{
				{
					FileName:       "./test_files/test1.txt",
					EncodedContent: encondeFileContent("This is a test"),
				},
				{
					FileName:       "./test_files/test2.txt",
					EncodedContent: encondeFileContent("This is another test"),
				},
			},
			expected: []*pb.ProcessedFile{
				{
					Processed: false,
				},
				{
					Processed: false,
				},
			},
		},
		"send_files": {
			input: []*pb.File{
				{
					FileName:       "./test_files/nonexistentfile1.txt",
					EncodedContent: encondeFileContent("This is a test"),
				},
				{
					FileName:       "./test_files/nonexistentfile2.txt",
					EncodedContent: encondeFileContent("This is a test"),
				},
				{
					FileName:       "./test_files/nonexistentfile1.txt",
					EncodedContent: encondeFileContent("This is a test"),
				},
			},
			expected: []*pb.ProcessedFile{
				{
					Processed: true,
					FileName:  "./test_files/nonexistentfile1.txt",
				},
				{
					Processed: true,
					FileName:  "./test_files/nonexistentfile2.txt",
				},
				{
					Processed: false,
				},
			},
		},
	}

	for scenario, testcase := range testcases {
		t.Run(scenario, func(t *testing.T) {
			filediff.CreateTestFile("This is a test", "test1.txt")
			filediff.CreateTestFile("This is another test", "test2.txt")

			out, err := client.SendFiles(ctx)

			for _, v := range testcase.input {
				if err := out.Send(v); err != nil {
					t.Errorf("Error while sending message: %v", err)
				}
			}

			if err := out.CloseSend(); err != nil {
				t.Errorf("Error closing stream: %v", err)
			}

			var outputs []*pb.ProcessedFile
			for {
				o, err := out.Recv()
				if errors.Is(err, io.EOF) {
					break
				}

				outputs = append(outputs, o)
			}

			assert.Nil(t, err)
			assert.Equal(t, len(testcase.expected), len(outputs))
			for i, o := range outputs {
				assert.Equal(t, testcase.expected[i].Processed, o.Processed)
				assert.Equal(t, testcase.expected[i].FileName, o.FileName)
				assert.Equal(t, testcase.expected[i].Error, o.Error)
			}

			deleteTestFiles()
		})
	}
}
