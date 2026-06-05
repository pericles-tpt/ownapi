package functions

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/pkg/errors"
)

const (
	goBinHTTPSourcePrefix = "https://go.dev/dl"
)

var (
	tmpTarDst string

	expTarHash  string
	untarFilter = []string{"bin", "pkg", "src", "go.env"}
)

type GoBinaryMetadata = []GoVersionMetadata

type GoVersionMetadata struct {
	Version string           `json:"version"`
	Stable  bool             `json:"stable"`
	Files   []GoFileMetadata `json:"files"`
}

type GoFileMetadata struct {
	Filename string `json:"filename"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Version  string `json:"version"`
	Sha256   string `json:"sha256"`
	Size     int64  `json:"size"`
	Kind     string `json:"kind"`
}

func validateUnpackGoTar(tarDestDir string) error {
	var (
		version    = strings.Split(runtime.Version(), " ")[0]
		_, err     = os.Stat(tmpTarDst)
		refetchTar = os.IsNotExist(err)

		buf bytes.Buffer
	)
	if refetchTar {
		// Download file from source
		fmt.Printf("Downloading %s binary...\n", version)
		buf, err = getGoTar(version, tmpTarDst)
		if err != nil {
			return errors.Wrapf(err, "error occurred getting go tar file from %s", goBinHTTPSourcePrefix)
		}
	} else {
		f, err := os.OpenFile(tmpTarDst, os.O_RDONLY, 0400)
		if err != nil {
			return errors.Wrapf(err, "failed to open tar file on disk %s", tmpTarDst)
		}
		_, err = buf.ReadFrom(f)
		if err != nil {
			return errors.Wrapf(err, "failed to read tar from open fp %s", tmpTarDst)
		}
	}

	// After download check hashes match
	// TODO: This accounts for 1s of slowness, the JSON payload is quite big but it doesn't look like you can
	// 		 retrieve metadata for a specific go version from the docs: https://pkg.go.dev/golang.org/x/website/internal/dl
	expTarHash, err = getRemoteGoBinaryTarHash(version)
	if err != nil {
		return errors.Wrapf(err, "error occurred getting go binary tar hash from %s", goBinHTTPSourcePrefix)
	}
	s := sha256.New()
	s.Write(buf.Bytes())
	gotTarHash := fmt.Sprintf("%x", s.Sum(nil))
	if gotTarHash != expTarHash {
		return fmt.Errorf("hash mismatch between got != exp, %s != %s", gotTarHash, expTarHash)
	}

	// Untar and retrieve binary
	gzr, err := gzip.NewReader(&buf)
	if err != nil {
		return errors.Wrap(err, "failed to untar go tar file")
	}
	defer gzr.Close()
	tr := tar.NewReader(gzr)
	for {
		var exit bool
		header, err := tr.Next()
		switch {
		case err == io.EOF:
			exit = true
		case err != nil:
			continue
		case header != nil:
			// bin
			name := header.Name
			var (
				found  bool
				filter string
			)
			for _, filter = range untarFilter {
				pre := fmt.Sprintf("go/%s", filter)
				if filter == name || strings.HasPrefix(name, pre) {
					found = true
					break
				}
			}

			path := fmt.Sprintf("%s/%s", tarDestDir, strings.TrimPrefix(name, "go/"))
			isBin := filter == "bin" || strings.HasPrefix(name, "go/pkg/tool")
			if found {
				if header.Typeflag == tar.TypeReg {
					buf := make([]byte, header.Size)
					_, err := io.ReadFull(tr, buf)
					if err != nil {
						return errors.Wrapf(err, "failed to read '%s' in go tar file", header.Name)
					}

					var mode os.FileMode = 0400
					if isBin {
						mode = 0700
					}
					err = os.WriteFile(path, buf, mode)
					if err != nil {
						return errors.Wrapf(err, "failed to write contents of '%s' to '%s'", header.Name, path)
					}
				} else {
					var mode os.FileMode = 0700
					if isBin {
						mode = 0700
					}
					err = os.MkdirAll(path, mode)
					if err != nil {
						return errors.Wrapf(err, "failed to mkdir at path: %s", path)
					}
				}
			}
		}

		if exit {
			break
		}
	}

	return nil
}

func getGoTar(version string, destPath string) (bytes.Buffer, error) {
	buf := bytes.Buffer{}
	basename := fmt.Sprintf("%s.%s-%s.tar.gz", version, runtime.GOOS, runtime.GOARCH)
	resp, err := http.Get(fmt.Sprintf("%s/%s", goBinHTTPSourcePrefix, basename))
	if err != nil {
		return buf, errors.Wrap(err, "failed to get go tar from HTTP source")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return buf, fmt.Errorf("https status not ok, code: %d, status: %s", resp.StatusCode, resp.Status)
	}

	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return buf, errors.Wrap(err, "failed to read body after successful request")
	}

	err = os.WriteFile(destPath, buf.Bytes(), 0666)
	if err != nil {
		return buf, errors.Wrapf(err, "failed to write retrieve bytes to file: %s", destPath)
	}

	return buf, nil
}

func getRemoteGoBinaryTarHash(version string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("%s/?mode=json&include=all", goBinHTTPSourcePrefix))
	if err != nil {
		return "", errors.Wrap(err, "failed to get metadata from HTTP source")
	}
	defer resp.Body.Close()

	buf := bytes.Buffer{}
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "failed to read body after successful request")
	}

	md := GoBinaryMetadata{}
	err = json.Unmarshal(buf.Bytes(), &md)
	if err != nil {
		return "", errors.Wrap(err, "failed to unmarshal body to GoBinaryMetadata")
	}

	var (
		arch = runtime.GOARCH
		os   = runtime.GOOS
	)
	for _, vmd := range md {
		if vmd.Version == version {
			for _, f := range vmd.Files {
				if f.Arch == arch && f.OS == os {
					return f.Sha256, nil
				}
			}
			return "", fmt.Errorf("found version %s in response but no matching arch and os found", version)
		}
	}
	return "", fmt.Errorf("no version matching %s found", version)
}
