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

func (s *fileServer) CompareFile(ctx context.Context, in *pb.File) (*pb.FileDiff, error) {
	isEqual, err := filediff.IsSameFile(in.FileName, in.EncodedContent)
	if err != nil {
		return &pb.FileDiff{}, err
	}

	return &pb.FileDiff{IsSame: isEqual}, nil
}
