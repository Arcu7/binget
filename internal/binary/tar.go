// TODO: Move this into its separate package/file
package binary

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"iter"
	"log/slog"
	"os"
)

type TarData struct {
	header *tar.Header
	reader io.Reader
}

// IMPORTANT: The caller is responsible for closing the returned *os.File
func (f *Finder) extractBinaryFromTarGz(file *os.File) (outFile *os.File, err error) {
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		f.logger.Error("Failed to create gzip reader", slog.String("error", err.Error()))
		return nil, err
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)

	for data, err := range tarIterator(tarReader) {
		if err == io.EOF {
			break
		}

		switch data.header.Typeflag {
		case tar.TypeDir:
			continue
		case tar.TypeReg:
			// Skip common non-executable files by name
			if data.header.Name == "LICENSE" || data.header.Name == "LICENSE.txt" {
				f.logger.Debug("Skipping license file in tar.gz archive", slog.String("file", data.header.Name))
				continue
			}
			if data.header.Name == "README" || data.header.Name == "README.md" {
				f.logger.Debug("Skipping readme file in tar.gz archive", slog.String("file", data.header.Name))
				continue
			}

			// Skip non-executable files by checking the mode
			if data.header.Mode&0o111 == 0 {
				f.logger.Debug("Skipping non-executable file in tar.gz archive", slog.String("file", data.header.Name))
				continue
			}

			outFile, err = os.Create(data.header.Name)
			if err != nil {
				return nil, err
			}

			_, err = io.Copy(outFile, data.reader)
			if err != nil {
				_ = outFile.Close()
				_ = os.Remove(outFile.Name())
				return nil, err
			}

			// Seek back to start so caller can read the file
			_, err = outFile.Seek(0, io.SeekStart)
			if err != nil {
				_ = outFile.Close()
				_ = os.Remove(outFile.Name())
				return nil, err
			}

			f.logger.Info("Found executable from tar.gz archive", slog.String("file", data.header.Name))
			return outFile, nil
		default:
			return nil, fmt.Errorf("unsupported tar entry type: %v", data.header.Typeflag)
		}
	}

	return nil, fmt.Errorf("no executable file found in tar.gz archive")
}

func tarIterator(tarReader *tar.Reader) iter.Seq2[TarData, error] {
	return func(yield func(TarData, error) bool) {
		for {
			hdr, err := tarReader.Next()
			if err == io.EOF {
				yield(TarData{}, io.EOF)
				return
			}
			if err != nil {
				yield(TarData{}, err)
				return
			}
			if !yield(TarData{header: hdr, reader: tarReader}, nil) {
				return
			}
		}
	}
}
