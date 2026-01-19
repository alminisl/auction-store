package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
)

var (
	ErrInvalidKey   = errors.New("invalid encryption key: must be 32 bytes (64 hex characters)")
	ErrInvalidNonce = errors.New("invalid nonce size")
	ErrDecryption   = errors.New("decryption failed")
)

// AESEncryptor handles AES-256-GCM encryption/decryption
type AESEncryptor struct {
	aead cipher.AEAD
}

// NewAESEncryptor creates a new AES-256-GCM encryptor from a hex-encoded key
func NewAESEncryptor(hexKey string) (*AESEncryptor, error) {
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, ErrInvalidKey
	}
	if len(key) != 32 {
		return nil, ErrInvalidKey
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &AESEncryptor{aead: aead}, nil
}

// Encrypt encrypts plaintext and returns ciphertext and nonce
func (e *AESEncryptor) Encrypt(plaintext []byte) (ciphertext []byte, nonce []byte, err error) {
	nonce = make([]byte, e.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}

	ciphertext = e.aead.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}

// Decrypt decrypts ciphertext using the provided nonce
func (e *AESEncryptor) Decrypt(ciphertext []byte, nonce []byte) ([]byte, error) {
	if len(nonce) != e.aead.NonceSize() {
		return nil, ErrInvalidNonce
	}

	plaintext, err := e.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, ErrDecryption
	}

	return plaintext, nil
}

// EncryptString encrypts a string and returns ciphertext and nonce
func (e *AESEncryptor) EncryptString(plaintext string) (ciphertext []byte, nonce []byte, err error) {
	return e.Encrypt([]byte(plaintext))
}

// DecryptString decrypts ciphertext and returns the plaintext string
func (e *AESEncryptor) DecryptString(ciphertext []byte, nonce []byte) (string, error) {
	plaintext, err := e.Decrypt(ciphertext, nonce)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// NonceSize returns the nonce size required for this encryptor
func (e *AESEncryptor) NonceSize() int {
	return e.aead.NonceSize()
}
