package server

import (
	"context"

	"github.com/aacuadras/ha-utils/lib/filediff"
	"github.com/aacuadras/ha-utils/server/pb"
)

type fileServer struct {
	pb.UnimplementedFileUtilsServer
}

func NewFileServer() pb.FileUtilsServer {
	return &fileServer{}
}

// This function receives the encoded contents of a configuration file and the desired path that the file should
// be in, if the contents of the files are the same as the ones currently in the path, then it will inform the client
// that the file was not processed and ignore it
func (s *fileServer) SendFile(ctx context.Context, in *pb.File) (*pb.ProcessedFile, error) {
	// Only substitute the file if it's not equal
	isEqual, err := filediff.IsSameFile(in.FileName, in.EncodedContent)
	if err != nil {
		return &pb.ProcessedFile{}, err
	}

	if !isEqual {
		if err := filediff.ReplaceFile(in.FileName, in.EncodedContent); err != nil {
			return &pb.ProcessedFile{
				Processed: false,
				Error:     err.Error(),
			}, nil
		}

		return &pb.ProcessedFile{
			Processed: true,
			FileName:  in.FileName,
		}, nil
	}

	return &pb.ProcessedFile{
		Processed: false,
	}, nil
}

// This function compares the encoded contents of a file with a file currently in the path provided, it will return
// if the current file has the same contents or if it's different
func (s *fileServer) CompareFile(ctx context.Context, in *pb.File) (*pb.FileDiff, error) {
	isEqual, err := filediff.IsSameFile(in.FileName, in.EncodedContent)
	if err != nil {
		return &pb.FileDiff{}, err
	}

	return &pb.FileDiff{IsSame: isEqual}, nil
}
