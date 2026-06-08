package binary

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/pericles-tpt/ownapi/config"
	"github.com/pericles-tpt/ownapi/secrets"
	"github.com/pericles-tpt/ownapi/utility"
	"github.com/pkg/errors"
)

var (
	root string

	names   = []string{}
	paths   = []string{}
	sources = []string{}

	ignoreExts = []string{"md", "txt"}
)

func Init(key []byte) error {
	root = config.GetDataDir("_bin")
	var (
		fd              *os.File
		oldFileContents = []byte{}
		hashesPath      = fmt.Sprintf("%s/hashes.txt", root)
	)
	st, err := os.Stat(hashesPath)
	if err != nil && !os.IsNotExist(err) {
		return errors.Wrapf(err, "failed to stat file that should exist: %s", hashesPath)
	} else if err == nil {
		fd, err = os.OpenFile(hashesPath, os.O_RDWR, 0666)
		if err != nil {
			return errors.Wrapf(err, "failed to open hashes file: %s", hashesPath)
		}

		oldFileContents = make([]byte, st.Size())
		_, err := fd.Read(oldFileContents)
		if err != nil {
			return errors.Wrapf(err, "failed to read hashes file: %s", hashesPath)
		}
		defer utility.WipeBytes(oldFileContents)

		err = loadSavedHashes(oldFileContents, key)
		if err != nil {
			return errors.Wrap(err, "failed to load binary hashes from hashes.txt, file might've been tampered with")
		}
	}

	filePaths, fileBasenames, fileContents, _, err := utility.WalkMaxDepth1(root, nil, func(s string) bool {
		var (
			parts  = strings.Split(s, ".")
			hasExt = len(parts) > 1
		)
		if hasExt {
			ext := parts[1]
			if _, ignore := utility.Contains(ext, ignoreExts); ignore {
				return false
			}
		}
		return true
	}, func(s string) bool { return s != ".quarrantine" })
	if err != nil {
		return errors.Wrapf(err, "failed to walk root: %s", root)
	}

	var (
		newNames   = make([]string, 0, len(fileBasenames))
		newPaths   = make([]string, 0, len(fileBasenames))
		newSources = make([]string, 0, len(fileBasenames))
		newHashes  = make([][]byte, 0, len(fileBasenames))

		gotHash []byte
	)
	for i, bn := range fileBasenames {
		utility.WipeBytes(gotHash)

		gotHash = utility.HashSHA256(fileContents[i])
		path := filePaths[i]

		if _, found := utility.Contains(bn, names); found {
			valid := ValidateBinary(root, bn, path, gotHash)
			if valid {
				// Make sure file is executable
				err := os.Chmod(path, 0700)
				if err != nil {
					return errors.Wrapf(err, "failed to make binary executable after validation: %s", path)
				}

				continue
			}

			return fmt.Errorf("failed to validate binary in hashes.txt: %s", bn)
		}

		source, hashStr, success := promptUserForSourceHash(root, bn, path, gotHash)
		if success {
			// Make sure file is executable
			err := os.Chmod(path, 0700)
			if err != nil {
				return errors.Wrapf(err, "failed to make binary executable after validation: %s", path)
			}

			newNames = append(newNames, bn)
			newPaths = append(newPaths, path)
			newSources = append(newSources, source)
			newHashes = append(newHashes, hashStr)
		}
	}
	utility.WipeBytes(gotHash)

	newFileContents := oldFileContents
	defer utility.WipeBytes(newFileContents)
	if len(newNames) > 0 {
		names = append(names, newNames...)
		paths = append(paths, newPaths...)
		sources = append(sources, newSources...)

		lines := make([]string, 0, len(newNames))
		for i, n := range newNames {
			secrets.AddKeyToLKKS(fmt.Sprintf("binary:%s", n), newHashes[i])
			lines = append(lines, fmt.Sprintf("%s %s %s %x", n, newSources[i], newPaths[i], newHashes[i]))
		}

		newFileContents = append(newFileContents, []byte(fmt.Sprintf("\n%s", strings.Join(lines, "\n")))...)
		encrypted, err := utility.EncryptAES256GCM(newFileContents, key)
		if err != nil {
			return errors.Wrap(err, "failed to encrypt hashes content to update file")
		}

		err = os.WriteFile(hashesPath, encrypted, 0666)
		if err != nil {
			return errors.Wrapf(err, "failed to write encrypted data to file: %s", hashesPath)
		}
	}

	return nil
}

func loadSavedHashes(encrypted []byte, key []byte) error {
	dec, err := utility.DecryptAES256GCM(encrypted, key)
	if err != nil {
		return errors.Wrap(err, "failed to decrypt bytes")
	}
	defer utility.WipeBytes(dec)

	// EXPECTING: NAME SOURCE_URL PATH HASH
	var (
		buf = make([]byte, 0, 50)
		at  = 0

		name   string
		path   string
		source string
	)
	for _, b := range dec {
		r := rune(b)
		if !unicode.IsSpace(r) {
			buf = append(buf, b)
			continue
		}

		if len(buf) > 0 {
			switch at {
			case 0:
				name = string(buf)
			case 1:
				source = string(buf)
			case 2:
				path = string(buf)
			case 3:
				// Add entry
				names = append(names, name)
				sources = append(sources, source)
				paths = append(paths, path)
				hash, err := hex.DecodeString(string(buf))
				if err != nil {
					return errors.Wrapf(err, "failed to decode string to []byte for binary '%s'", name)
				}
				err = secrets.AddKeyToLKKS(fmt.Sprintf("binary:%s", name), hash)
				if err != nil {
					return errors.Wrapf(err, "failed to add hash for binary '%s' to LKKS", name)
				}
			}

			buf = buf[:0]
			at = (at + 1) % 4
		}
	}
	if at == 3 && len(buf) > 0 {
		names = append(names, name)
		sources = append(sources, source)
		paths = append(paths, path)
		hash, err := hex.DecodeString(string(buf))
		if err != nil {
			return errors.Wrapf(err, "failed to decode string to []byte for binary '%s'", name)
		}
		err = secrets.AddKeyToLKKS(fmt.Sprintf("binary:%s", name), hash)
		if err != nil {
			return errors.Wrapf(err, "failed to add hash for binary '%s' to LKKS", name)
		}

		at = (at + 1) % 4
	}

	if at != 0 {
		return fmt.Errorf("invalid space/newline separated arguments in file, expected: n %% 4 == 0, got: n %% 4 == %d", at)
	}

	return nil
}

func promptUserForSourceHash(root, name, path string, wantHash []byte) (string, []byte, bool) {
	var (
		source, hashStr string
	)

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Printf("Found new binary %s, provide a source (just hit ENTER for none): ", name)
	scanner.Scan()
	// TODO: Validate source, try to retrieve file from there
	source = scanner.Text()
	if source == "" {
		source = "NONE"
	}

	fmt.Printf("\nProvide the hash you EXPECT the binary to have: ")
	scanner.Scan()
	// TODO: Validate source, try to retrieve file from there
	hashStr = scanner.Text()
	hashStr = strings.TrimSpace(hashStr)
	gotHash, err := hex.DecodeString(hashStr)
	if err != nil {
		fmt.Printf("Failed to decode string: %s\n", err)
		tryMoveToQuarrantine(root, name, path, nil, &hashStr)
		return source, gotHash, false
	}

	if !utility.BytesEqual(gotHash, wantHash) {
		tryMoveToQuarrantine(root, name, path, nil, &hashStr)
		return source, gotHash, false
	}

	fmt.Printf("Hashes match, added %s!\n", name)

	return source, gotHash, true
}

func tryMoveToQuarrantine(root, name, path string, hashBytes *[]byte, hashString *string) {
	var got string
	if hashBytes != nil {
		got = fmt.Sprintf("%x", *hashBytes)
	} else if hashString != nil {
		got = *hashString
	}

	quarrantineName := fmt.Sprintf("%s_%d_%s", name, time.Now().UnixMilli(), got)
	quarrantinePath := fmt.Sprintf("%s/.quarrantine/%s", root, quarrantineName)

	fmt.Printf("[BINARY] binary %s doesn't match expected hash, moving to: %s\n", name, quarrantinePath)

	err := os.Rename(path, quarrantinePath)
	if err != nil {
		fmt.Printf("[BINARY] failed to move binary to quarrantine, attempting to remove")
		os.Remove(path)
	}
}
