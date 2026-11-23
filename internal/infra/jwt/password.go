package jwtutil

import (
	"crypto/rand"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"strings"

	"golang.org/x/crypto/argon2"
)

// HashPassword hashes a raw plaintext password using SHA-512 pre-hash + Argon2id.
// A fresh random salt of saltLength bytes is generated per call.
// Returned format: "base64(salt)$base64(hash)".
//
// Recommended saltLength: 32.
func HashPassword(password string, saltLength int) string {
	preHashed := hashSHA512(password)

	saltBytes := make([]byte, saltLength)
	if _, err := rand.Read(saltBytes); err != nil {
		panic("salt generation failed: " + err.Error())
	}

	// Argon2id: time=1, memory=64MB, threads=4, keyLen=32 — conservative, solid defaults
	hash := argon2.IDKey([]byte(preHashed), saltBytes, 1, 64*1024, 4, 32)

	return base64.RawStdEncoding.EncodeToString(saltBytes) + "$" +
		base64.RawStdEncoding.EncodeToString(hash)
}

// Compare verifies a raw plaintext password against a stored hash string.
// Uses constant-time comparison to prevent timing side-channel attacks.
func Compare(password, storedHash string) bool {
	parts := strings.Split(storedHash, "$")
	if len(parts) != 2 {
		return false
	}

	salt, err1 := base64.RawStdEncoding.DecodeString(parts[0])
	expectedHash, err2 := base64.RawStdEncoding.DecodeString(parts[1])
	if err1 != nil || err2 != nil {
		return false
	}

	preHashed := hashSHA512(password)
	computedHash := argon2.IDKey([]byte(preHashed), salt, 1, 64*1024, 4, 32)

	return subtle.ConstantTimeCompare(computedHash, expectedHash) == 1
}

func hashSHA512(s string) string {
	h := sha512.Sum512([]byte(s))
	return hex.EncodeToString(h[:])
}
