package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/pericles-tpt/ownapi/utility"
	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

const (
	SECRET_NAME_SIZE_LIMIT = 50
	SECRETS_PREFIX         = "secret:"
)

var (
	separator       = rune(':')
	keyringId       int
	keynameToIdSize = map[string]struct {
		Id   int
		Size int
	}{}

	// TODO: Figure out how to change this to the "process" keyring
	keyringType       = unix.KEY_SPEC_USER_KEYRING
	keyringTypeString = "user"
)

func Init(secretsFile string, pipelinesBytes []byte) error {
	pw, err := setupSecretsOrPromptPassword(secretsFile)
	if err != nil {
		return err
	}
	defer utility.WipeBytes(pw)

	existingSecretKeys, err := loadToLKKS(secretsFile, pw)
	if err != nil {
		return err
	}

	err = promptForMissingSecrets(existingSecretKeys, pipelinesBytes, secretsFile, pw)
	if err != nil {
		return err
	}

	return nil
}

func promptForMissingSecrets(existingKeys []string, pipelinesFileBytes []byte, secretsFileName string, pw []byte) error {
	pipelinesFileString := string(pipelinesFileBytes)
	pipelineSecrets, _, _ := GetSecretsOffsetsLens(pipelinesFileString)

	var (
		newSecretsBuffer = make([]byte, 0, 1024)
	)
	defer utility.WipeBytes(newSecretsBuffer)
	for _, secretKey := range pipelineSecrets {
		if _, exists := utility.Contains(secretKey, existingKeys); !exists {
			var (
				validInput bool
				quit       bool
			)
			fmt.Printf("Missing secret '%s', paste the secret here to securely store it OR type 'q' to continue without it: \n", secretKey)
			var (
				secretValue = []byte{}
				err         error
			)
			for !validInput {
				secretValue, err = term.ReadPassword(int(os.Stdin.Fd()))
				if err != nil {
					return errors.Wrapf(err, "failed to read user input to store new secret for '%s'", secretKey)
				}
				defer utility.WipeBytes(secretValue)

				validInput = true
				secretLen := len(secretValue)
				if secretLen == 0 {
					fmt.Println("No secret provided, provide secret again:")
					validInput = false
				}
				if secretLen == 1 && rune(secretValue[0]) == rune('q') {
					fmt.Println("Quitting...")
					utility.WipeBytes(secretValue)
					quit = true
					break
				}

				var confirmLen string
				fmt.Printf("Provided secret is %d characters long, is that correct? [y/N]\n", secretLen)
				fmt.Scanln(&confirmLen)
				isYes := (len(confirmLen) == 1 && strings.ToLower(string(confirmLen[0])) == "y")
				if !isYes {
					fmt.Println("NOT YES!")
					fmt.Println("Secret doesn't match length expected by user, provide secret again:")
					validInput = false
				}
			}

			newSecretsBuffer = append(newSecretsBuffer, []byte(secretKey)...)
			newSecretsBuffer = append(newSecretsBuffer, byte(':'))
			newSecretsBuffer = append(newSecretsBuffer, secretValue...)
			newSecretsBuffer = append(newSecretsBuffer, byte('\n'))

			utility.WipeBytes(secretValue)

			if quit {
				break
			}
		}
	}

	if len(newSecretsBuffer) > 0 {
		err := writeNewSecretsToFileAndLKKS(newSecretsBuffer, secretsFileName, pw)
		if err != nil {
			return errors.Wrap(err, "failed to write new secrets from user prompts to file")
		}
	}

	return nil
}

// GetSecretsOffsetsLens, retrieves the secret name, offsets and len
//
// NOTE: `offset` and `len` are calculated AFTER the `SECRETS_PREFIX` and colon
func GetSecretsOffsetsLens(contents string) ([]string, []int, []int) {
	var (
		secretsPrefixIdx int

		collectSecretName bool
		secretName        = make([]rune, 0, SECRET_NAME_SIZE_LIMIT)

		pipelineSecretNames   = make([]string, 0, 50)
		pipelineSecretOffsets = make([]int, 0, 50)
		pipelineSecretLens    = make([]int, 0, 50)
	)
	for i, c := range contents {
		if collectSecretName {
			if strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_-", c) {
				secretName = append(secretName, c)
			} else {
				pipelineSecretNames = append(pipelineSecretNames, string(secretName))
				pipelineSecretLens = append(pipelineSecretLens, len(secretName))

				secretName = make([]rune, 0, 50)
				collectSecretName = false
			}
		} else if c == rune(SECRETS_PREFIX[secretsPrefixIdx]) {
			if secretsPrefixIdx == len(SECRETS_PREFIX)-1 {
				collectSecretName = true
				secretsPrefixIdx = -1
				pipelineSecretOffsets = append(pipelineSecretOffsets, i)
			}
			secretsPrefixIdx++
		} else {
			secretsPrefixIdx = 0
		}
	}

	// Post loop case in case secret is at the end of the string
	if collectSecretName {
		pipelineSecretNames = append(pipelineSecretNames, string(secretName))
		pipelineSecretLens = append(pipelineSecretLens, len(secretName))
	}

	return pipelineSecretNames, pipelineSecretOffsets, pipelineSecretLens
}

func MaybeReplaceSecretsInString(target string) (bool, string, error) {
	maybeSecretNames, _, _ := GetSecretsOffsetsLens(target)
	if len(maybeSecretNames) > 0 {
		newStringParts := make([]string, 0, len(maybeSecretNames))

		var (
			replaceSecretIdxs  = make([]int, 0, len(maybeSecretNames))
			replaceSecretBytes = make([][]byte, 0, len(maybeSecretNames))
		)
		defer func(toWipe [][]byte) {
			for i, bs := range toWipe {
				for j := range bs {
					toWipe[i][j] = 0
				}
			}
		}(replaceSecretBytes)

		for _, sn := range maybeSecretNames {
			fullSecretName := fmt.Sprintf("%s%s", SECRETS_PREFIX, sn)
			partsAroundSecret := strings.Split(target, fullSecretName)

			newStringParts = append(newStringParts, partsAroundSecret[0])

			secret, err := getKeyFromLKKS(sn)
			if err != nil {
				return false, "", errors.Wrapf(err, "failed to retrieve secret '%s' from LKKS", sn)
			}
			replaceSecretIdxs = append(replaceSecretIdxs, len(newStringParts))
			replaceSecretBytes = append(replaceSecretBytes, secret)

			newStringParts = append(newStringParts, "PLACEHOLDER")

			if len(partsAroundSecret) > 1 {
				target = partsAroundSecret[1]
			}
		}

		for i, idx := range replaceSecretIdxs {
			newStringParts[idx] = string(replaceSecretBytes[i])
		}

		target = strings.Join(newStringParts, "")
	}

	return len(maybeSecretNames) > 0, target, nil
}

func setupSecretsOrPromptPassword(filename string) ([]byte, error) {
	var (
		pw  []byte
		err error

		newPassword    bool
		pwPromptPrefix = "Secrets file found, please provide your secrets password:"
	)
	_, err = os.Stat(filename)
	if err != nil {
		if !os.IsNotExist(err) {
			return pw, err
		}
		pwPromptPrefix = "Secrets file NOT found, please provide your NEW secrets password:"
		newPassword = true
	}

	fmt.Println(pwPromptPrefix)
	var validPW bool
	for !validPW {
		pw, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return pw, err
		}
		validPW = true

		if newPassword {
			// TODO: Add length, complexity, rainbow table, etc checks here in a loop
			fmt.Println("Please re-enter your password:")
			pwC, err := term.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				return pw, err
			}

			passwordsMatch := utility.BytesEqual(pw, pwC)
			if !passwordsMatch {
				fmt.Println("Password mismatch, please enter your password again:")
			}
			validPW = passwordsMatch
		}
	}

	hash := sha256.New()
	hash.Write(pw)
	pw = hash.Sum(nil)

	if newPassword {
		// Create file with a sample encrypted contents, test decryption
		contents := []byte("sample:abc\n")
		encrypted, err := encryptGCM(contents, pw)
		if err != nil {
			return pw, errors.Wrap(err, "failed to encrypt sample file contents with new pw")
		}
		decrypted, err := decryptGCM(encrypted, pw)
		if err != nil {
			return pw, errors.Wrap(err, "failed to decrypt sample file contents with new pw")
		}
		if !utility.BytesEqual(contents, decrypted) {
			return pw, errors.Wrap(err, "failed to match decrypted contents with original sample file contents for new pw")
		}
		err = os.WriteFile(filename, encrypted, 0644)
		if err != nil {
			return pw, errors.Wrapf(err, "failed to create initial secrets file at '%s'", filename)
		}
	}

	return pw, nil
}

func loadToLKKS(file string, pw []byte) ([]string, error) {
	var (
		keys []string
		err  error
	)
	keyringId, err = unix.KeyctlGetKeyringID(keyringType, true)
	if err != nil {
		return keys, errors.Wrap(err, "failed to get keyring for storage of secrets")
	}

	fb, err := os.ReadFile(file)
	if err != nil {
		return keys, err
	}
	defer utility.WipeBytes(fb)

	dfb, err := decryptGCM(fb, pw)
	if err != nil {
		return keys, err
	}
	defer utility.WipeBytes(dfb)

	var (
		k string

		v []byte
		i int
	)
	for i < len(dfb) {
		k, v, i, err = consumeNextSecret(dfb, i)
		if err != nil {
			return keys, errors.Wrap(err, "failed to consume next secret")
		}
		defer utility.WipeBytes(v)

		keys = append(keys, k)

		err = addKeyToLKKS(k, v)
		if err != nil {
			return keys, errors.Wrapf(err, "failed to store key '%s' in LKKS", k)
		}
		utility.WipeBytes(v)
	}

	return keys, nil
}

func consumeNextSecret(bs []byte, startIdx int) (string, []byte, int, error) {
	var (
		keyParts = make([]rune, 0, len(bs)/2)
		value    = make([]byte, 0, len(bs)/2)
		i        int
	)
	if startIdx >= len(bs) {
		return "", value, i, fmt.Errorf("provided `startIdx` out of range of bytes: %d >= %d", startIdx, len(bs))
	}

	var (
		r               rune
		beforeSeparator = true
		err             error
		size            int
	)
	defer utility.WipeBytesOnErr(value, &err)
	for i = startIdx; i < len(bs); i += size {
		r, size = utf8.DecodeRune(bs[i:])
		invalid := r == utf8.RuneError && size == 1
		if invalid {
			err = fmt.Errorf("invalid rune found at idx: %d, of decrypted file", i)
			return "", value, i, err
		}

		if r == separator {
			beforeSeparator = false
			continue
		} else if r == rune('\n') {
			break
		}

		if beforeSeparator {
			keyParts = append(keyParts, r)
		} else {
			tmpBs := make([]byte, size)
			utf8.EncodeRune(tmpBs, r)
			value = append(value, tmpBs...)
			utility.WipeBytes(tmpBs)
		}
	}

	// breaks before newline must do `i + 1`
	return string(keyParts), value, i + 1, nil
}

func addKeyToLKKS(k string, v []byte) error {
	keyId, err := unix.AddKey(keyringTypeString, k, v, keyringId)
	if err != nil {
		return errors.Wrapf(err, "failed to save secret with key: '%s', to keyring", k)
	}
	keynameToIdSize[k] = struct {
		Id   int
		Size int
	}{
		Id:   keyId,
		Size: len(v),
	}

	return nil
}

// getKeyFromLKKS, it's the caller's responsibility to zero bytes
//
// after use and avoid converting to less safe types like string
func getKeyFromLKKS(k string) ([]byte, error) {
	vt, err := getKeyValue(k)
	if err != nil {
		return vt, errors.Wrapf(err, "failed to get secret with key: '%s'", k)
	}
	return vt, nil
}

func writeNewSecretsToFileAndLKKS(newSecrets []byte, file string, pw []byte) error {
	fb, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	defer utility.WipeBytes(fb)

	dfb, err := decryptGCM(fb, pw)
	if err != nil {
		return err
	}
	defer utility.WipeBytes(dfb)

	dfb = append(dfb, newSecrets...)

	efb, err := encryptGCM(dfb, pw)
	if err != nil {
		return err
	}
	defer utility.WipeBytes(efb)

	return os.WriteFile(file, efb, 0644)
}

func getKeyValue(keyName string) ([]byte, error) {
	props := keynameToIdSize[keyName]

	// Allocate buffer and read payload
	buf := make([]byte, props.Size)
	n, err := unix.KeyctlBuffer(unix.KEYCTL_READ, props.Id, buf, 0)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

// SOURCE: https://dev.to/shrsv/encryption-and-decryption-in-go-a-hands-on-guide-3bcl
// TODO: Review these later
func encryptGCM(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func decryptGCM(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
