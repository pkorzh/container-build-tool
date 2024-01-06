package archive

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Compression int

const (
	Uncompressed Compression = iota
	Bzip2
	Gzip
)

func DetectCompression(source []byte) Compression {
	for compression, m := range map[Compression][]byte{
		Bzip2: {0x42, 0x5A, 0x68},
		Gzip:  {0x1F, 0x8B, 0x08},
	} {
		if len(source) < len(m) {
			continue
		}
		if bytes.Equal(m, source[:len(m)]) {
			return compression
		}
	}
	return Uncompressed
}

func DecompressStream(src io.Reader) (io.Reader, Compression, error) {
	buffer := bufio.NewReader(src)
	sig, err := buffer.Peek(10)

	if err != nil {
		return nil, Uncompressed, err
	}

	compression := DetectCompression(sig)

	switch compression {
	case Uncompressed:
		return buffer, Uncompressed, nil
	case Bzip2:
		return bzip2.NewReader(buffer), Bzip2, err
	case Gzip:
		gzipReader, err := gzip.NewReader(buffer)
		if err != nil {
			return nil, Gzip, err
		}
		return gzipReader, Gzip, nil
	default:
		return nil, Uncompressed, fmt.Errorf("unsupported compression: %d", compression)
	}
}

func CompressStream(dst io.WriteCloser, compression Compression) (io.WriteCloser, error) {
	switch compression {
	case Uncompressed:
		return dst, nil
	case Bzip2:
		return nil, fmt.Errorf("bzip2 compression not supported")
	case Gzip:
		return gzip.NewWriter(dst), nil
	default:
		return nil, fmt.Errorf("unsupported compression: %d", compression)
	}
}

func Tar(src string, compression Compression) (io.ReadCloser, error) {
	pipeReader, pipeWriter := io.Pipe()

	compressed, err := CompressStream(pipeWriter, compression)
	if err != nil {
		return nil, err
	}

	go func() {
		tr := tar.NewWriter(compressed)

		defer func() {
			tr.Close()
			compressed.Close()
			pipeWriter.Close()
		}()

		if err := filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return fmt.Errorf("walk: %w", err)
			}

			relPath, err := filepath.Rel(src, path)
			if err != nil {
				return fmt.Errorf("rel: %w", err)
			}

			if relPath == "." {
				return nil
			}

			fi, err := os.Lstat(path)
			if err != nil {
				return fmt.Errorf("lstat: %w", err)
			}

			var link string
			if fi.Mode()&os.ModeSymlink != 0 {
				link, err = os.Readlink(path)
				if err != nil {
					return fmt.Errorf("readlink: %w", err)
				}
			}

			header, err := tar.FileInfoHeader(fi, link)
			if err != nil {
				return fmt.Errorf("header: %w", err)
			}

			header.Name = relPath

			if err := tr.WriteHeader(header); err != nil {
				return fmt.Errorf("write header: %w", err)
			}

			if header.Typeflag == tar.TypeReg {
				file, err := os.Open(path)
				if err != nil {
					return fmt.Errorf("open: %w", err)
				}

				_, err = io.Copy(tr, file)
				file.Close()
				if err != nil {
					return fmt.Errorf("copy: %w", err)
				}
			}

			return nil

		}); err != nil {
			fmt.Println(err)
		}
	}()

	return pipeReader, nil
}

func Untar(src io.Reader, dst string) error {
	decompressed, _, err := DecompressStream(src)
	if err != nil {
		return err
	}

	tr := tar.NewReader(decompressed)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar read: %w", err)
		}

		header.Name = filepath.Clean(header.Name)

		path := filepath.Join(dst, header.Name)
		fi := header.FileInfo()
		mask := fi.Mode()

		switch header.Typeflag {
		case tar.TypeDir:
			if fi, err := os.Lstat(path); !(err == nil && fi.IsDir()) {
				if err := os.MkdirAll(path, mask); err != nil {
					return fmt.Errorf("mkdir: %w", err)
				}
			}
		case tar.TypeReg:
			file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, mask)
			if err != nil {
				return fmt.Errorf("open: %w", err)
			}
			if _, err := io.Copy(file, tr); err != nil {
				file.Close()
				return fmt.Errorf("copy: %w", err)
			}
			file.Close()
		case tar.TypeSymlink:
			targetPath := filepath.Join(filepath.Dir(path), header.Linkname)
			if !strings.HasPrefix(targetPath, dst) {
				return fmt.Errorf("symlink: %w", err)
			}
			if err := os.Symlink(header.Linkname, path); err != nil {
				return fmt.Errorf("symlink: %w", err)
			}
		default:
			return fmt.Errorf("unsupported type: %d", header.Typeflag)
		}
	}

	return nil
}

func IsArchivePath(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	reader, _, err := DecompressStream(file)
	if err != nil {
		return false
	}

	tr := tar.NewReader(reader)
	_, err = tr.Next()
	if err == io.EOF {
		return false
	}
	if err != nil {
		return false
	}

	return true
}
