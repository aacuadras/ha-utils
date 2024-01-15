package server

import (
	"context"
	"io"
	"sync"

	"github.com/aacuadras/ha-utils/lib/filediff"
	"github.com/aacuadras/ha-utils/server/pb"
)

type fileServer struct {
	pb.UnimplementedFileUtilsServer
	mu             sync.Mutex
	fileDiffs      []*pb.FileDiff
	processedFiles []*pb.ProcessedFile
}

func NewFileServer() pb.FileUtilsServer {
	return &fileServer{}
}

// This function receives the encoded contents of a configuration file and the desired path that the file should
// be in, if the contents of the files are the same as the ones currently in the path, then it will inform the client
// that the file was not processed and ignore it
func (s *fileServer) SendFile(ctx context.Context, in *pb.File) (*pb.ProcessedFile, error) {
	// Only substitute the file if it's not equal
	var isEqual bool
	var err error
	if filediff.DoesFileExist(in.FileName) {
		isEqual, err = filediff.IsSameFile(in.FileName, in.EncodedContent)
		if err != nil {
			return &pb.ProcessedFile{}, err
		}
	} else {
		isEqual = false
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

// This function works the same way as SendFile, but it receives a stream of File so it can process multiple files
// instead of one at a time
func (s *fileServer) SendFiles(stream pb.FileUtils_SendFilesServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		var isEqual bool
		if filediff.DoesFileExist(in.FileName) {
			isEqual, err = filediff.IsSameFile(in.FileName, in.EncodedContent)
			if err != nil {
				return err
			}
		} else {
			isEqual = false
		}

		s.mu.Lock()
		if !isEqual {
			if err := filediff.ReplaceFile(in.FileName, in.EncodedContent); err != nil {
				s.processedFiles = append(s.processedFiles, &pb.ProcessedFile{
					Processed: false,
					Error:     err.Error(),
				})
			} else {
				s.processedFiles = append(s.processedFiles, &pb.ProcessedFile{
					Processed: true,
					FileName:  in.FileName,
				})
			}
		} else {
			s.processedFiles = append(s.processedFiles, &pb.ProcessedFile{
				Processed: false,
			})
		}

		rn := make([]*pb.ProcessedFile, len(s.processedFiles))
		copy(rn, s.processedFiles)
		s.mu.Unlock()

		for _, note := range rn {
			if err := stream.Send(note); err != nil {
				return err
			}
			s.processedFiles = s.processedFiles[:0]
		}
	}
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

// This function works the same way as CompareFile, but it receives a stream of File so it can process multiple files
// instead of one at a time
func (s *fileServer) CompareFiles(stream pb.FileUtils_CompareFilesServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		isEqual, err := filediff.IsSameFile(in.FileName, in.EncodedContent)
		if err != nil {
			return err
		}

		s.mu.Lock()
		s.fileDiffs = append(s.fileDiffs, &pb.FileDiff{
			IsSame: isEqual,
		})
		rn := make([]*pb.FileDiff, len(s.fileDiffs))
		copy(rn, s.fileDiffs)
		s.mu.Unlock()

		for _, note := range rn {
			if err := stream.Send(note); err != nil {
				return err
			}
			s.fileDiffs = s.fileDiffs[:0]
		}
	}
}
