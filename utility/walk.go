package utility

import (
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"
)

func WalkMaxDepth1(root string, fileModMap *map[string]time.Time, fileFilter, dirFilter func(s string) bool) ([]string, []string, [][]byte, *[]bool, error) {
	var (
		filePaths     = []string{}
		fileBasenames = []string{}
		filesContents = [][]byte{}
		isNew         *[]bool
	)
	dirents, err := os.ReadDir(root)
	if err != nil {
		return filePaths, fileBasenames, filesContents, isNew, errors.Wrapf(err, "failed to read `%s`", root)
	}

	filePaths = make([]string, 0, len(dirents))
	fileBasenames = make([]string, 0, len(dirents))
	filesContents = make([][]byte, 0, len(dirents))
	if fileModMap != nil {
		in := make([]bool, 0, len(dirents))
		isNew = &in
	}

	var (
		parents = []struct {
			lastChild int
			path      string
		}{
			{
				lastChild: len(dirents) - 1,
				path:      root,
			},
		}
		currParentIdx = 0
	)

	for i := 0; i < len(dirents); i++ {
		if i > parents[currParentIdx].lastChild {
			currParentIdx++
		}
		currParentPath := parents[currParentIdx].path
		currParentLastChild := parents[currParentIdx].lastChild

		var (
			de   = dirents[i]
			name = de.Name()
			path = fmt.Sprintf("%s/%s", currParentPath, name)
		)
		if de.IsDir() {
			if dirFilter(name) {
				nestedDirents, err := os.ReadDir(path)
				if err != nil {
					return filePaths, fileBasenames, filesContents, isNew, errors.Wrapf(err, "failed to read dir in root `%s`", path)
				}

				// NOTE: Only support 1 level of file nesting from root, currently
				nestedFiles := make([]os.DirEntry, 0, len(nestedDirents))
				for _, de := range nestedDirents {
					if de.Type().IsRegular() {
						nestedFiles = append(nestedFiles, de)
					}
				}
				dirents = append(dirents, nestedFiles...)

				parents = append(parents, struct {
					lastChild int
					path      string
				}{
					lastChild: currParentLastChild + len(nestedDirents),
					path:      path,
				})
			}
			continue
		}

		if de.Type().IsRegular() && fileFilter(name) {
			if fileModMap != nil {
				info, err := de.Info()
				if err != nil {
					fmt.Printf("WARN: Failed to read file: %s\n", name)
					continue
				}
				currLastModified := info.ModTime()

				(*isNew) = append((*isNew), false)
				var (
					prevLastModified time.Time
					exists           bool
				)
				if prevLastModified, exists = (*fileModMap)[name]; !exists {
					(*isNew)[len(*isNew)-1] = true
				}
				isModified := prevLastModified != currLastModified
				(*fileModMap)[name] = currLastModified

				if !isModified {
					continue
				}
			}

			var contents []byte
			contents, err := os.ReadFile(path)
			if err != nil {
				return filePaths, fileBasenames, filesContents, isNew, errors.Wrapf(err, "failed to read file '%s'", path)
			}

			filesContents = append(filesContents, contents)
			filePaths = append(filePaths, path)
			fileBasenames = append(fileBasenames, name)
		}
	}

	return filePaths, fileBasenames, filesContents, isNew, nil
}
