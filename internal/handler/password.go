package handler

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
	"strings"
)

const (
	passwordHashAlgorithm  = "pbkdf2_sha256"
	passwordHashIterations = 120000
	passwordSaltBytes      = 16
	passwordKeyBytes       = 32
)

func hashPassword(password string) (string, error) {
	salt := make([]byte, passwordSaltBytes)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}

	key := pbkdf2SHA256([]byte(password), salt, passwordHashIterations, passwordKeyBytes)
	return strings.Join([]string{
		passwordHashAlgorithm,
		strconv.Itoa(passwordHashIterations),
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	}, "$"), nil
}

func verifyPassword(encoded, password string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 || parts[0] != passwordHashAlgorithm {
		return false
	}

	iterations, err := strconv.Atoi(parts[1])
	if err != nil || iterations <= 0 {
		return false
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil {
		return false
	}

	expected, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return false
	}

	actual := pbkdf2SHA256([]byte(password), salt, iterations, len(expected))
	return subtle.ConstantTimeCompare(actual, expected) == 1
}

func pbkdf2SHA256(password, salt []byte, iterations, keyLen int) []byte {
	hashLen := sha256.Size
	blockCount := (keyLen + hashLen - 1) / hashLen
	derived := make([]byte, 0, blockCount*hashLen)

	for block := 1; block <= blockCount; block++ {
		mac := hmac.New(sha256.New, password)
		mac.Write(salt)

		var blockIndex [4]byte
		binary.BigEndian.PutUint32(blockIndex[:], uint32(block))
		mac.Write(blockIndex[:])

		u := mac.Sum(nil)
		t := make([]byte, len(u))
		copy(t, u)

		for i := 1; i < iterations; i++ {
			mac = hmac.New(sha256.New, password)
			mac.Write(u)
			u = mac.Sum(nil)

			for j := range t {
				t[j] ^= u[j]
			}
		}

		derived = append(derived, t...)
	}

	return derived[:keyLen]
}
