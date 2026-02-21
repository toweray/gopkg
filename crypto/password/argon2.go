package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

var (
	ErrInvalidHash         = errors.New("invalid hash format")
	ErrIncompatibleVersion = errors.New("incompatible argon2 version")
)

// Hasher is the interface for hashing and verifying passwords.
// Implementations must be safe for concurrent use.
type Hasher interface {
	Hash(password string) (string, error)
	Verify(password, hash string) (bool, error)
}

// Argon2Config holds parameters for the Argon2id algorithm.
type Argon2Config struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// DefaultArgon2Config returns secure default parameters for Argon2id.
func DefaultArgon2Config() Argon2Config {
	return Argon2Config{
		Memory:      64 * 1024,
		Iterations:  3,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	}
}

// Argon2Hasher implements Hasher using the Argon2id algorithm.
type Argon2Hasher struct {
	cfg Argon2Config
}

// NewArgon2 returns a new Argon2Hasher with the given config.
func NewArgon2(cfg Argon2Config) *Argon2Hasher {
	return &Argon2Hasher{cfg: cfg}
}

// Hash hashes password using Argon2id and returns an encoded string.
// Format: $argon2id$v=19$m=65536,t=3,p=2$<salt>$<hash>
func (h *Argon2Hasher) Hash(password string) (string, error) {
	salt := make([]byte, h.cfg.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		h.cfg.Iterations,
		h.cfg.Memory,
		h.cfg.Parallelism,
		h.cfg.KeyLength,
	)

	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		h.cfg.Memory,
		h.cfg.Iterations,
		h.cfg.Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

// Verify checks whether password matches the Argon2id encoded hash.
func (h *Argon2Hasher) Verify(password, encoded string) (bool, error) {
	params, salt, hash, err := h.decodeHash(encoded)
	if err != nil {
		return false, err
	}

	computed := argon2.IDKey(
		[]byte(password),
		salt,
		params.iterations,
		params.memory,
		params.parallelism,
		params.keyLength,
	)

	return subtle.ConstantTimeCompare(hash, computed) == 1, nil
}

type argon2Params struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	keyLength   uint32
}

// decodeHash parses an encoded Argon2id hash.
// Format: $argon2id$v=19$m=65536,t=3,p=2$<salt>$<hash>
func (h *Argon2Hasher) decodeHash(encoded string) (*argon2Params, []byte, []byte, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 {
		return nil, nil, nil, ErrInvalidHash
	}

	if parts[1] != "argon2id" {
		return nil, nil, nil, ErrInvalidHash
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return nil, nil, nil, ErrInvalidHash
	}
	if version != argon2.Version {
		return nil, nil, nil, ErrIncompatibleVersion
	}

	var params argon2Params
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d",
		&params.memory, &params.iterations, &params.parallelism,
	); err != nil {
		return nil, nil, nil, ErrInvalidHash
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, nil, ErrInvalidHash
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, nil, ErrInvalidHash
	}

	params.keyLength = uint32(len(hash))

	return &params, salt, hash, nil
}
