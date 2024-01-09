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

func (s *fileServer) CompareFile(ctx context.Context, in *pb.File) (*pb.FileDiff, error) {
	decodedContents, err := filediff.DecodeFile(in.EncodedContent)
	if err != nil {
		return &pb.FileDiff{}, err
	}

	isEqual, err := filediff.IsSameFile(in.FileName, decodedContents)
	if err != nil {
		return &pb.FileDiff{}, err
	}

	return &pb.FileDiff{IsSame: isEqual}, nil
}
